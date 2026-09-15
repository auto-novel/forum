export interface PostListReturn {
  path: string;
  scrollTop: number;
}

export let postListReturn: PostListReturn | undefined;

export function setPostListReturn(value?: PostListReturn) {
  postListReturn = value;
}
