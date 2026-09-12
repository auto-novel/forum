import { ref } from 'vue';

export interface AppNotification {
  id: number;
  type: 'success' | 'error';
  title: string;
  description: string;
}

export const notifications = ref<AppNotification[]>([]);

let nextNotificationId = 0;

function addNotification(
  type: AppNotification['type'],
  title: string,
  description: string,
) {
  notifications.value.unshift({
    id: ++nextNotificationId,
    type,
    title,
    description,
  });
}

export function notifySuccess(description: string) {
  addNotification('success', '操作成功', description);
}

export function notifyError(description: string) {
  addNotification('error', '操作失败', description);
}

export function dismissNotification(id: number) {
  const index = notifications.value.findIndex(
    (notification) => notification.id === id,
  );
  if (index >= 0) notifications.value.splice(index, 1);
}
