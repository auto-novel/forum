import { onBeforeUnmount, toValue, watch, type MaybeRefOrGetter } from 'vue';
import { onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router';

const CONFIRM_MESSAGE = '有尚未保存的修改，确定要离开吗？';

export function useUnsavedChangesGuard(
  hasUnsavedChanges: MaybeRefOrGetter<boolean>,
) {
  let listening = false;

  function handleBeforeUnload(event: BeforeUnloadEvent) {
    if (!toValue(hasUnsavedChanges)) return;
    event.preventDefault();
    event.returnValue = '';
  }

  function setListening(value: boolean) {
    if (value === listening) return;
    listening = value;
    if (value) window.addEventListener('beforeunload', handleBeforeUnload);
    else window.removeEventListener('beforeunload', handleBeforeUnload);
  }

  function confirmNavigation() {
    return !toValue(hasUnsavedChanges) || window.confirm(CONFIRM_MESSAGE);
  }

  watch(() => toValue(hasUnsavedChanges), setListening, {
    immediate: true,
    flush: 'sync',
  });
  onBeforeUnmount(() => setListening(false));
  onBeforeRouteLeave(confirmNavigation);
  onBeforeRouteUpdate(confirmNavigation);
}
