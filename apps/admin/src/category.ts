const categoryDisplay = [
  { slug: 'announcements', title: '站务公告' },
  { slug: 'novel', title: '小说讨论' },
  { slug: 'feedback', title: '意见反馈' },
];

export function categoryTitle(slug: string): string {
  return categoryDisplay.find((item) => item.slug === slug)?.title ?? slug;
}

export function categoryOrder(slug: string): number {
  const index = categoryDisplay.findIndex((item) => item.slug === slug);
  return index < 0 ? categoryDisplay.length : index;
}
