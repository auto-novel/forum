import type { InjectionKey, Ref } from 'vue';

export interface MobileNavigation {
  open: Ref<boolean>;
}

export const mobileNavigationKey: InjectionKey<MobileNavigation> =
  Symbol('mobile-navigation');
