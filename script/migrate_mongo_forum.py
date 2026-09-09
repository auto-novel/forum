#!/usr/bin/env python3
"""Migrate auto-novel articles and comments directly into forum PostgreSQL."""

from __future__ import annotations

import argparse
import json
import os
import sys
import tempfile
from collections import Counter
from collections.abc import Iterator
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

try:
    from bson import ObjectId
    from bson.errors import InvalidId
    from pymongo import MongoClient
    from pymongo.database import Database
    from pymongo.errors import PyMongoError
    import psycopg
except ImportError as error:
    raise SystemExit(
        "缺少迁移依赖，请先执行："
        "python3 -m pip install -r script/requirements-migration.txt"
    ) from error


CATEGORY_SLUGS = {"General": "novel", "Guide": "guide", "Support": "feedback"}
POST_SUBJECT = 0
NOVEL_SUBJECT = 1
PUBLISHED = 0
HIDDEN = 1


class MigrationError(RuntimeError):
    pass


def value_detail(value: Any) -> str:
    rendered = repr(value)
    if len(rendered) > 200:
        rendered = rendered[:197] + "..."
    return f"类型 {type(value).__name__}，值 {rendered}"


@dataclass
class SourceStats:
    users: int = 0
    articles: int = 0
    comments: int = 0
    skipped_articles: int = 0
    skipped_comments: int = 0
    missing_auth_users: set[str] = field(default_factory=set)
    migratable_article_ids: set[str] = field(default_factory=set)
    migratable_root_comment_ids: set[str] = field(default_factory=set)
    article_categories: Counter[str] = field(default_factory=Counter)
    article_statuses: Counter[int] = field(default_factory=Counter)
    comment_statuses: Counter[int] = field(default_factory=Counter)
    comment_subjects: Counter[int] = field(default_factory=Counter)
    comment_levels: Counter[str] = field(default_factory=Counter)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="将 auto-novel MongoDB 的 article/comment-alt 迁移到 forum"
    )
    parser.add_argument(
        "--mongo-uri",
        default=os.getenv("MONGO_URI"),
        help="MongoDB URI；默认读取 MONGO_URI",
    )
    parser.add_argument(
        "--mongo-db",
        default=os.getenv("MONGO_DB", "main"),
        help="MongoDB 数据库名，默认 main",
    )
    parser.add_argument(
        "--forum-dsn",
        default=os.getenv("FORUM_DATABASE_URL"),
        help="forum PostgreSQL DSN；默认读取 FORUM_DATABASE_URL",
    )
    parser.add_argument(
        "--auth-dsn",
        default=os.getenv("AUTH_DATABASE_URL"),
        help="auth PostgreSQL DSN；默认读取 AUTH_DATABASE_URL",
    )
    parser.add_argument(
        "--batch-size", type=int, default=1_000, help="流式读取批次大小，默认 1000"
    )
    parser.add_argument(
        "--mapping-file",
        help="ID 映射 JSONL 输出路径；默认在系统临时目录创建",
    )
    parser.add_argument(
        "--execute",
        action="store_true",
        help="通过全部预检后执行迁移；不指定时只预检",
    )
    args = parser.parse_args()
    missing = [
        name
        for name, value in (
            ("MONGO_URI/--mongo-uri", args.mongo_uri),
            ("FORUM_DATABASE_URL/--forum-dsn", args.forum_dsn),
            ("AUTH_DATABASE_URL/--auth-dsn", args.auth_dsn),
        )
        if not value
    ]
    if missing:
        parser.error("缺少连接参数：" + "、".join(missing))
    if args.batch_size <= 0:
        parser.error("--batch-size 必须大于 0")
    return args


def mongo_batches(
    database: Database,
    name: str,
    batch_size: int,
    query: dict[str, Any] | None = None,
    projection: dict[str, int] | None = None,
) -> Iterator[list[dict]]:
    cursor = (
        database[name]
        .find(query or {}, projection)
        .sort("_id", 1)
        .batch_size(batch_size)
    )
    batch: list[dict] = []
    try:
        for document in cursor:
            batch.append(document)
            if len(batch) == batch_size:
                yield batch
                batch = []
        if batch:
            yield batch
    finally:
        cursor.close()


def object_id(value: Any, field: str) -> str:
    result = "" if value is None else str(value)
    if not result:
        raise MigrationError(f"{field} 不能为空：{value_detail(value)}")
    return result


def bson_object_id(value: str, field: str) -> ObjectId:
    try:
        return ObjectId(value)
    except (InvalidId, TypeError) as error:
        raise MigrationError(
            f"{field} 不是有效的 Mongo ObjectId：{value_detail(value)}"
        ) from error


def text(value: Any, field: str, max_length: int | None = None) -> str:
    if not isinstance(value, str):
        raise MigrationError(f"{field} 必须是字符串：{value_detail(value)}")
    if "\x00" in value:
        raise MigrationError(
            f"{field} 包含 PostgreSQL 不支持的 NUL 字符：{value_detail(value)}"
        )
    if max_length is not None and len(value) > max_length:
        raise MigrationError(
            f"{field} 长度 {len(value)} 超过限制 {max_length}："
            f"{value_detail(value)}"
        )
    return value


def boolean(value: Any, field: str) -> bool:
    if not isinstance(value, bool):
        raise MigrationError(f"{field} 必须是布尔值：{value_detail(value)}")
    return value


def pg_integer(value: Any, field: str) -> int:
    if isinstance(value, bool) or not isinstance(value, int):
        raise MigrationError(f"{field} 必须是整数：{value_detail(value)}")
    if value < 0 or value > 2_147_483_647:
        raise MigrationError(
            f"{field} 超出 PostgreSQL int 非负数范围：{value_detail(value)}"
        )
    return value


def instant(value: Any, field: str) -> datetime:
    if isinstance(value, str):
        normalized = value[:-1] + "+00:00" if value.endswith("Z") else value
        try:
            value = datetime.fromisoformat(normalized)
        except ValueError as error:
            raise MigrationError(
                f"{field} 不是日期：{value_detail(value)}"
            ) from error
    if not isinstance(value, datetime):
        raise MigrationError(f"{field} 不是日期：{value_detail(value)}")
    return value.replace(tzinfo=timezone.utc) if value.tzinfo is None else value


def article_active_at(article: dict) -> datetime:
    article_id = object_id(article.get("_id"), "article._id")
    change_at = article.get("changeAt")
    if change_at is None:
        return instant(article.get("updateAt"), f"article[{article_id}].updateAt")
    return instant(change_at, f"article[{article_id}].changeAt")


def create_mapping_file(requested_path: str | None) -> tuple[Path, Any]:
    if requested_path:
        path = Path(requested_path).expanduser().resolve()
        handle = path.open("x", encoding="utf-8")
        return path, handle
    descriptor, raw_path = tempfile.mkstemp(
        prefix="forum-migration-id-map-", suffix=".jsonl"
    )
    os.close(descriptor)
    path = Path(raw_path)
    return path, path.open("w", encoding="utf-8")


def write_mapping(handle: Any, entity: str, source_id: str, target_id: int) -> None:
    json.dump(
        {"type": entity, "sourceId": source_id, "targetId": target_id},
        handle,
        ensure_ascii=False,
        separators=(",", ":"),
    )
    handle.write("\n")


def load_auth_users(cursor: psycopg.Cursor, usernames: set[str]) -> dict[str, int]:
    cursor.execute(
        "select id, username from auth_user where username = any(%s)",
        (sorted(usernames),),
    )
    return {username: user_id for user_id, username in cursor.fetchall()}


def resolve_authors(
    database: Database,
    auth_cursor: psycopg.Cursor,
    documents: list[dict],
    entity: str,
    missing_auth_users: set[str] | None = None,
) -> dict[str, tuple[int, str]]:
    source_values: dict[str, Any] = {}
    for document in documents:
        document_id = object_id(document.get("_id"), f"{entity}._id")
        user = document.get("user")
        source_values[object_id(user, f"{entity}[{document_id}].user")] = user

    source_users: dict[str, str] = {}
    for user in database["user"].find(
        {"_id": {"$in": list(source_values.values())}},
        {"_id": 1, "username": 1},
    ):
        user_id = object_id(user.get("_id"), "user._id")
        source_users[user_id] = text(
            user.get("username"), f"user[{user_id}].username", 128
        )
    missing = sorted(source_values.keys() - source_users.keys())
    if missing:
        raise MigrationError(
            f"{entity} 引用了不存在的 Mongo 用户：{'、'.join(missing[:20])}"
        )

    auth_users = load_auth_users(auth_cursor, set(source_users.values()))
    missing = sorted(set(source_users.values()) - auth_users.keys())
    if missing_auth_users is not None:
        missing_auth_users.update(missing)
    return {
        source_id: (auth_users[username], username)
        for source_id, username in source_users.items()
        if username in auth_users
    }


def validate_articles(
    database: Database,
    auth_cursor: psycopg.Cursor,
    batch_size: int,
    stats: SourceStats,
) -> None:
    projection = {
        "_id": 1,
        "title": 1,
        "content": 1,
        "category": 1,
        "hidden": 1,
        "locked": 1,
        "pinned": 1,
        "numViews": 1,
        "createAt": 1,
        "updateAt": 1,
        "changeAt": 1,
        "user": 1,
    }
    for articles in mongo_batches(
        database, "article", batch_size, projection=projection
    ):
        authors = resolve_authors(
            database,
            auth_cursor,
            articles,
            "article",
            stats.missing_auth_users,
        )
        for article in articles:
            article_id = object_id(article.get("_id"), "article._id")
            if object_id(
                article.get("user"), f"article[{article_id}].user"
            ) not in authors:
                stats.skipped_articles += 1
                continue
            text(article.get("title"), f"article[{article_id}].title", 500)
            text(article.get("content"), f"article[{article_id}].content")
            category = article.get("category")
            if category not in CATEGORY_SLUGS:
                raise MigrationError(
                    f"article[{article_id}] 存在未知分类：{value_detail(category)}"
                )
            hidden = boolean(
                article.get("hidden", False), f"article[{article_id}].hidden"
            )
            boolean(article.get("locked", False), f"article[{article_id}].locked")
            boolean(article.get("pinned", False), f"article[{article_id}].pinned")
            pg_integer(article.get("numViews", 0), f"article[{article_id}].numViews")
            for field in ("createAt", "updateAt"):
                instant(article.get(field), f"article[{article_id}].{field}")
            article_active_at(article)
            stats.migratable_article_ids.add(article_id)
            stats.article_categories[CATEGORY_SLUGS[category]] += 1
            stats.article_statuses[HIDDEN if hidden else PUBLISHED] += 1
            stats.articles += 1


def validate_comments(
    database: Database,
    auth_cursor: psycopg.Cursor,
    batch_size: int,
    stats: SourceStats,
) -> None:
    projection = {
        "_id": 1,
        "site": 1,
        "content": 1,
        "hidden": 1,
        "createAt": 1,
        "parent": 1,
        "user": 1,
    }
    for query in ({"parent": None}, {"parent": {"$ne": None}}):
        for comments in mongo_batches(
            database,
            "comment-alt",
            batch_size,
            query=query,
            projection=projection,
        ):
            authors = resolve_authors(
                database,
                auth_cursor,
                comments,
                "comment",
                stats.missing_auth_users,
            )
            article_ids: set[str] = set()
            parent_values: dict[str, Any] = {}
            sites: dict[str, str] = {}
            candidates: list[dict] = []
            for comment in comments:
                comment_id = object_id(comment.get("_id"), "comment._id")
                if object_id(
                    comment.get("user"), f"comment[{comment_id}].user"
                ) not in authors:
                    stats.skipped_comments += 1
                    continue
                site = text(
                    comment.get("site"), f"comment[{comment_id}].site", 255
                )
                if not site:
                    raise MigrationError(
                        f"comment[{comment_id}].site 不能为空：{value_detail(site)}"
                    )
                sites[comment_id] = site
                text(comment.get("content"), f"comment[{comment_id}].content")
                instant(comment.get("createAt"), f"comment[{comment_id}].createAt")
                boolean(
                    comment.get("hidden", False),
                    f"comment[{comment_id}].hidden",
                )
                if site.startswith("article-"):
                    article_ids.add(site.removeprefix("article-"))
                if comment.get("parent") is not None:
                    parent = comment["parent"]
                    parent_values[
                        object_id(parent, f"comment[{comment_id}].parent")
                    ] = parent
                candidates.append(comment)

            existing_articles = {
                str(value["_id"])
                for value in database["article"].find(
                    {
                        "_id": {
                            "$in": [
                                bson_object_id(value, "comment.site")
                                for value in article_ids
                            ]
                        }
                    },
                    {"_id": 1},
                )
            }
            missing = sorted(article_ids - existing_articles)
            if missing:
                raise MigrationError(
                    f"评论引用了不存在的文章：{'、'.join(missing[:20])}"
                )

            parents = {
                str(value["_id"]): value
                for value in database["comment-alt"].find(
                    {"_id": {"$in": list(parent_values.values())}},
                    {"_id": 1, "parent": 1, "site": 1},
                )
            }
            missing = sorted(parent_values.keys() - parents.keys())
            if missing:
                raise MigrationError(
                    f"评论引用了不存在的父评论：{'、'.join(missing[:20])}"
                )
            for comment in candidates:
                comment_id = str(comment["_id"])
                site = sites[comment_id]
                if (
                    site.startswith("article-")
                    and site.removeprefix("article-")
                    not in stats.migratable_article_ids
                ):
                    stats.skipped_comments += 1
                    continue
                parent = comment.get("parent")
                if (
                    parent is not None
                    and str(parent) not in stats.migratable_root_comment_ids
                ):
                    stats.skipped_comments += 1
                    continue
                if parent is not None:
                    parent_id = str(parent)
                    parent_comment = parents[parent_id]
                    if parent_comment.get("parent") is not None:
                        raise MigrationError(
                            f"comment[{comment_id}] 的父评论 {parent_id} "
                            "不是一级评论："
                            f"parent={value_detail(parent_comment.get('parent'))}"
                        )
                    if parent_comment.get("site") != site:
                        raise MigrationError(
                            f"comment[{comment_id}] 与父评论 {parent_id} "
                            "不属于同一主体："
                            f"comment.site={value_detail(site)}，"
                            "parent.site="
                            f"{value_detail(parent_comment.get('site'))}"
                        )

                hidden = comment.get("hidden", False)
                stats.comment_statuses[HIDDEN if hidden else PUBLISHED] += 1
                stats.comment_subjects[
                    POST_SUBJECT if site.startswith("article-") else NOVEL_SUBJECT
                ] += 1
                if parent is None:
                    stats.comment_levels["root"] += 1
                    stats.migratable_root_comment_ids.add(comment_id)
                else:
                    stats.comment_levels["reply"] += 1
                stats.comments += 1


def validate_source(
    database: Database, auth_dsn: str, batch_size: int
) -> SourceStats:
    stats = SourceStats()
    with psycopg.connect(auth_dsn) as connection:
        with connection.cursor() as cursor:
            validate_articles(database, cursor, batch_size, stats)
            validate_comments(database, cursor, batch_size, stats)
    stats.users = database["user"].count_documents({})
    return stats


def preflight_target(forum_dsn: str) -> dict[str, int]:
    with psycopg.connect(forum_dsn) as connection:
        with connection.cursor() as cursor:
            cursor.execute(
                """
                select data_type, character_maximum_length
                from information_schema.columns
                where table_schema = 'public' and table_name = 'comment'
                  and column_name = 'subject_key'
                """
            )
            subject_key_definition = cursor.fetchone()
            if subject_key_definition != ("character varying", 255):
                raise MigrationError(
                    "forum.comment.subject_key 必须是 varchar(255)，"
                    f"当前为 {subject_key_definition!r}"
                )
            cursor.execute(
                "select (select count(*) from post), "
                "(select count(*) from comment)"
            )
            post_count, comment_count = cursor.fetchone()
            if post_count or comment_count:
                raise MigrationError(
                    f"目标表必须为空，当前 post={post_count}、comment={comment_count}"
                )
            cursor.execute(
                "select slug, id from category where slug = any(%s)",
                (list(CATEGORY_SLUGS.values()),),
            )
            category_ids = dict(cursor.fetchall())
    missing = sorted(set(CATEGORY_SLUGS.values()) - category_ids.keys())
    if missing:
        raise MigrationError("forum 缺少分类：" + "、".join(missing))
    return category_ids


def load_id_map(
    cursor: psycopg.Cursor, table: str, source_ids: set[str]
) -> dict[str, int]:
    if not source_ids:
        return {}
    if table not in {"migration_article_map", "migration_comment_map"}:
        raise ValueError(f"unsupported migration map: {table}")
    cursor.execute(
        f"select source_id, target_id from {table} where source_id = any(%s)",
        (list(source_ids),),
    )
    result = dict(cursor.fetchall())
    missing = sorted(source_ids - result.keys())
    if missing:
        raise MigrationError(f"{table} 缺少 ID 映射：{'、'.join(missing[:20])}")
    return result


def insert_article(
    cursor: psycopg.Cursor,
    article: dict,
    author: tuple[int, str],
    category_ids: dict[str, int],
) -> int:
    source_id = str(article["_id"])
    author_id, username = author
    cursor.execute(
        """
        insert into post (
            category_id, title, author_id, author_username, content, status,
            views_count, comments_count, comments_locked, pin_order,
            created_at, updated_at, active_at
        ) values (%s, %s, %s, %s, %s, %s, %s, 0, %s, %s, %s, %s, %s)
        returning id
        """,
        (
            category_ids[CATEGORY_SLUGS[article["category"]]],
            article["title"],
            author_id,
            username,
            article["content"],
            HIDDEN if article.get("hidden", False) else PUBLISHED,
            article.get("numViews", 0),
            article.get("locked", False),
            0 if article.get("pinned", False) else None,
            instant(article["createAt"], f"article[{source_id}].createAt"),
            instant(article["updateAt"], f"article[{source_id}].updateAt"),
            article_active_at(article),
        ),
    )
    return cursor.fetchone()[0]


def query_counts(cursor: psycopg.Cursor, query: str) -> dict[Any, int]:
    cursor.execute(query)
    return {key: count for key, count in cursor.fetchall()}


def require_counts(name: str, expected: Counter, actual: dict[Any, int]) -> None:
    if dict(expected) != actual:
        raise MigrationError(f"迁移后{name}统计不一致：{actual} != {dict(expected)}")


def validate_migrated(cursor: psycopg.Cursor, stats: SourceStats) -> None:
    require_counts(
        "文章分类",
        stats.article_categories,
        query_counts(
            cursor,
            "select c.slug, count(*) from post p join category c "
            "on c.id = p.category_id group by c.slug",
        ),
    )
    require_counts(
        "文章状态",
        stats.article_statuses,
        query_counts(cursor, "select status, count(*) from post group by status"),
    )
    require_counts(
        "评论状态",
        stats.comment_statuses,
        query_counts(cursor, "select status, count(*) from comment group by status"),
    )
    require_counts(
        "评论主体",
        stats.comment_subjects,
        query_counts(
            cursor, "select subject_type, count(*) from comment group by subject_type"
        ),
    )
    require_counts(
        "评论层级",
        stats.comment_levels,
        query_counts(
            cursor,
            "select case when root_id is null then 'root' else 'reply' end, "
            "count(*) from comment group by 1",
        ),
    )
    cursor.execute(
        """
        select m.source_id, p.id, p.comments_count,
               m.expected_published_comments
        from migration_article_map m
        join post p on p.id = m.target_id
        where p.comments_count != m.expected_published_comments
        limit 20
        """
    )
    comment_count_mismatches = cursor.fetchall()
    if comment_count_mismatches:
        raise MigrationError(
            "迁移后存在帖子公开评论计数不一致："
            f"(source_id, post_id, 实际值, 预期值)={comment_count_mismatches!r}"
        )
    cursor.execute(
        """
        select reply.id, reply.root_id, root.id, root.root_id,
               reply.subject_type, root.subject_type,
               reply.subject_key, root.subject_key
        from comment reply
        left join comment root on root.id = reply.root_id
        where reply.root_id is not null
          and (root.id is null or root.root_id is not null
               or root.subject_type != reply.subject_type
               or root.subject_key != reply.subject_key)
        limit 20
        """
    )
    invalid_relationships = cursor.fetchall()
    if invalid_relationships:
        raise MigrationError(
            "迁移后存在无效的评论父子关系："
            "(reply_id, root_id, 实际 root_id, root.root_id, "
            "reply.subject_type, root.subject_type, "
            "reply.subject_key, root.subject_key)="
            f"{invalid_relationships!r}"
        )


def migrate(
    database: Database,
    forum_dsn: str,
    auth_dsn: str,
    category_ids: dict[str, int],
    stats: SourceStats,
    batch_size: int,
    requested_mapping_path: str | None,
) -> tuple[int, int, Path]:
    article_count = 0
    comment_count = 0
    mapping_path, mapping_handle = create_mapping_file(requested_mapping_path)
    try:
        with psycopg.connect(auth_dsn) as auth_connection:
            with auth_connection.cursor() as auth_cursor:
                with psycopg.connect(forum_dsn) as connection:
                    with connection.transaction():
                        with connection.cursor() as cursor:
                            cursor.execute(
                                "lock table post, comment in share row exclusive mode"
                            )
                            cursor.execute(
                                "select (select count(*) from post), "
                                "(select count(*) from comment)"
                            )
                            target_counts = cursor.fetchone()
                            if target_counts != (0, 0):
                                raise MigrationError(
                                    "目标表在预检后发生变化："
                                    f"当前 post={target_counts[0]}、"
                                    f"comment={target_counts[1]}"
                                )
                            cursor.execute(
                                """
                                create temporary table migration_article_map (
                                    source_id text primary key,
                                    target_id bigint not null unique,
                                    expected_published_comments int not null default 0
                                ) on commit drop
                                """
                            )
                            cursor.execute(
                                """
                                create temporary table migration_comment_map (
                                    source_id text primary key,
                                    target_id bigint not null unique
                                ) on commit drop
                                """
                            )

                            for articles in mongo_batches(
                                database, "article", batch_size
                            ):
                                authors = resolve_authors(
                                    database, auth_cursor, articles, "article"
                                )
                                mappings: list[tuple[str, int]] = []
                                for article in articles:
                                    source_id = str(article["_id"])
                                    if (
                                        source_id
                                        not in stats.migratable_article_ids
                                    ):
                                        continue
                                    target_id = insert_article(
                                        cursor,
                                        article,
                                        authors[str(article["user"])],
                                        category_ids,
                                    )
                                    mappings.append((source_id, target_id))
                                    write_mapping(
                                        mapping_handle, "post", source_id, target_id
                                    )
                                if mappings:
                                    cursor.executemany(
                                        "insert into migration_article_map "
                                        "(source_id, target_id) values (%s, %s)",
                                        mappings,
                                    )
                                article_count += len(mappings)

                            def migrate_comments(query: dict[str, Any]) -> int:
                                migrated = 0
                                for comments in mongo_batches(
                                    database, "comment-alt", batch_size, query=query
                                ):
                                    authors = resolve_authors(
                                        database, auth_cursor, comments, "comment"
                                    )
                                    comments = [
                                        comment
                                        for comment in comments
                                        if str(comment["user"]) in authors
                                        and (
                                            not comment["site"].startswith(
                                                "article-"
                                            )
                                            or comment["site"].removeprefix(
                                                "article-"
                                            )
                                            in stats.migratable_article_ids
                                        )
                                        and (
                                            comment.get("parent") is None
                                            or str(comment["parent"])
                                            in stats.migratable_root_comment_ids
                                        )
                                    ]
                                    article_sources = {
                                        value["site"].removeprefix("article-")
                                        for value in comments
                                        if value["site"].startswith("article-")
                                    }
                                    article_map = load_id_map(
                                        cursor,
                                        "migration_article_map",
                                        article_sources,
                                    )
                                    parent_sources = {
                                        str(value["parent"])
                                        for value in comments
                                        if value.get("parent") is not None
                                    }
                                    parent_map = load_id_map(
                                        cursor,
                                        "migration_comment_map",
                                        parent_sources,
                                    )
                                    published_counts: Counter[str] = Counter()
                                    mappings: list[tuple[str, int]] = []
                                    for comment in comments:
                                        source_id = str(comment["_id"])
                                        author_id, username = authors[
                                            str(comment["user"])
                                        ]
                                        site = comment["site"]
                                        if site.startswith("article-"):
                                            article_source = site.removeprefix(
                                                "article-"
                                            )
                                            subject_type = POST_SUBJECT
                                            subject_key = str(
                                                article_map[article_source]
                                            )
                                            if not comment.get("hidden", False):
                                                published_counts[article_source] += 1
                                        else:
                                            subject_type = NOVEL_SUBJECT
                                            subject_key = site
                                        parent = comment.get("parent")
                                        root_id = (
                                            parent_map[str(parent)]
                                            if parent is not None
                                            else None
                                        )
                                        created_at = instant(
                                            comment["createAt"],
                                            f"comment[{source_id}].createAt",
                                        )
                                        cursor.execute(
                                            """
                                            insert into comment (
                                                subject_type, subject_key, root_id,
                                                content, author_id, author_username,
                                                status, created_at, updated_at
                                            ) values (
                                                %s, %s, %s, %s, %s, %s, %s, %s, %s
                                            ) returning id
                                            """,
                                            (
                                                subject_type,
                                                subject_key,
                                                root_id,
                                                comment["content"],
                                                author_id,
                                                username,
                                                HIDDEN
                                                if comment.get("hidden", False)
                                                else PUBLISHED,
                                                created_at,
                                                created_at,
                                            ),
                                        )
                                        target_id = cursor.fetchone()[0]
                                        mappings.append((source_id, target_id))
                                        write_mapping(
                                            mapping_handle,
                                            "comment",
                                            source_id,
                                            target_id,
                                        )
                                    if mappings:
                                        cursor.executemany(
                                            "insert into migration_comment_map "
                                            "(source_id, target_id) values (%s, %s)",
                                            mappings,
                                        )
                                    if published_counts:
                                        cursor.executemany(
                                            "update migration_article_map set "
                                            "expected_published_comments = "
                                            "expected_published_comments + %s "
                                            "where source_id = %s",
                                            [
                                                (count, source_id)
                                                for source_id, count
                                                in published_counts.items()
                                            ],
                                        )
                                    migrated += len(comments)
                                return migrated

                            comment_count += migrate_comments({"parent": None})
                            comment_count += migrate_comments(
                                {"parent": {"$ne": None}}
                            )
                            cursor.execute(
                                """
                                update post p set comments_count = (
                                    select count(*)::int from comment c
                                    where c.subject_type = %s
                                      and c.subject_key = p.id::text
                                      and c.status = %s
                                )
                                """,
                                (POST_SUBJECT, PUBLISHED),
                            )
                            if article_count != stats.articles:
                                raise MigrationError(
                                    "迁移后文章总数不一致："
                                    f"实际 {article_count}，预期 {stats.articles}"
                                )
                            if comment_count != stats.comments:
                                raise MigrationError(
                                    "迁移后评论总数不一致："
                                    f"实际 {comment_count}，预期 {stats.comments}"
                                )
                            validate_migrated(cursor, stats)
                            mapping_handle.flush()
                            os.fsync(mapping_handle.fileno())
    except BaseException:
        mapping_handle.close()
        mapping_path.unlink(missing_ok=True)
        raise
    try:
        mapping_handle.close()
    except OSError as error:
        print(f"警告：ID 映射文件关闭失败：{error}", file=sys.stderr)
    return article_count, comment_count, mapping_path


def main() -> int:
    args = parse_args()
    try:
        with MongoClient(args.mongo_uri, tz_aware=True) as client:
            database = client[args.mongo_db]
            stats = validate_source(
                database, args.auth_dsn, args.batch_size
            )
            category_ids = preflight_target(args.forum_dsn)
            print(
                f"预检通过：Mongo 用户 {stats.users}，"
                f"文章 {stats.articles}，评论 {stats.comments}"
            )
            if stats.missing_auth_users:
                missing_users = "、".join(
                    repr(username)
                    for username in sorted(stats.missing_auth_users)[:20]
                )
                print(
                    f"警告：auth 缺少用户 {missing_users}；"
                    f"将跳过文章 {stats.skipped_articles}、"
                    f"评论 {stats.skipped_comments}",
                    file=sys.stderr,
                )
            if not args.execute:
                print("当前为预检模式；确认结果后添加 --execute 执行迁移")
                return 0
            article_count, comment_count, mapping_path = migrate(
                database,
                args.forum_dsn,
                args.auth_dsn,
                category_ids,
                stats,
                args.batch_size,
                args.mapping_file,
            )
        print(
            f"迁移完成：文章 {article_count}，评论 {comment_count}；"
            f"ID 映射：{mapping_path}"
        )
        return 0
    except (MigrationError, PyMongoError, psycopg.Error, OSError) as error:
        print(f"迁移失败：{error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
