import type { InjectionKey, Ref } from 'vue';

export interface PostListMobileNavigation {
  open: Ref<boolean>;
}

export const postListMobileNavigationKey: InjectionKey<PostListMobileNavigation> =
  Symbol('post-list-mobile-navigation');
