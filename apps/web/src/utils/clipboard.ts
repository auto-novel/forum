/** 复制文本到剪贴板，优先异步剪贴板接口，失败时回退到选区复制。 */
export async function copyText(text: string): Promise<void> {
  const clipboard = globalThis.navigator?.clipboard;
  if (clipboard?.writeText) {
    try {
      await clipboard.writeText(text);
      return;
    } catch {
      // 非安全上下文或权限被拒绝时，继续尝试选区复制。
    }
  }
  if (!copyTextBySelection(text)) throw new Error('复制到剪贴板失败');
}

function copyTextBySelection(text: string): boolean {
  if (typeof document === 'undefined') return false;
  const area = document.createElement('textarea');
  area.value = text;
  area.setAttribute('readonly', '');
  area.style.position = 'fixed';
  area.style.top = '0';
  area.style.opacity = '0';
  document.body.appendChild(area);
  try {
    area.select();
    return document.execCommand?.('copy') === true;
  } catch {
    return false;
  } finally {
    area.remove();
  }
}
