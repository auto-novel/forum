import { computed, readonly, ref } from 'vue';

type Theme = 'light' | 'dark';

const STORAGE_KEY = 'forum:theme';
const TRANSITION_DURATION = 200;
const currentTheme = ref<Theme>('light');
let initialized = false;
let transitionTimer: number | undefined;

function storedTheme(): Theme | undefined {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value === 'light' || value === 'dark' ? value : undefined;
  } catch {
    return undefined;
  }
}

function preferredTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

function applyTheme(theme: Theme, animated = false) {
  const root = document.documentElement;

  if (animated) {
    window.clearTimeout(transitionTimer);
    root.classList.add('theme-transition');
    void root.offsetWidth;
  }

  root.dataset.theme = theme;
  document
    .querySelector<HTMLMetaElement>('meta[name="theme-color"]')
    ?.setAttribute('content', theme === 'dark' ? '#101014' : '#f7f7f8');

  if (animated) {
    transitionTimer = window.setTimeout(() => {
      root.classList.remove('theme-transition');
      transitionTimer = undefined;
    }, TRANSITION_DURATION);
  }
}

export function initializeTheme() {
  if (initialized) return;
  initialized = true;
  currentTheme.value = storedTheme() ?? preferredTheme();
  applyTheme(currentTheme.value);
}

export function useTheme() {
  initializeTheme();

  const isDark = computed(() => currentTheme.value === 'dark');

  function toggleTheme() {
    currentTheme.value = isDark.value ? 'light' : 'dark';
    applyTheme(currentTheme.value, true);
    try {
      localStorage.setItem(STORAGE_KEY, currentTheme.value);
    } catch {
      // The in-memory theme still works when storage is unavailable.
    }
  }

  return {
    theme: readonly(currentTheme),
    isDark,
    toggleTheme,
  };
}
