import { ref } from 'vue';

export interface AppNotification {
  id: number;
  type: 'success' | 'error';
  message: string;
}

export const notifications = ref<AppNotification[]>([]);

let nextNotificationId = 0;

function addNotification(type: AppNotification['type'], message: string) {
  notifications.value.unshift({
    id: ++nextNotificationId,
    type,
    message,
  });
}

export function notifySuccess(message: string) {
  addNotification('success', message);
}

export function notifyError(message: string) {
  addNotification('error', message);
}

export function dismissNotification(id: number) {
  const index = notifications.value.findIndex(
    (notification) => notification.id === id,
  );
  if (index >= 0) notifications.value.splice(index, 1);
}
