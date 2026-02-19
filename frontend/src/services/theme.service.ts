import { ref } from 'vue';

const isDarkMode = ref(false);

export function useTheme() {
  const toggleTheme = () => {
    isDarkMode.value = !isDarkMode.value;
    const root = document.getElementsByTagName('html')[0];
    if (isDarkMode.value) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
    localStorage.setItem('theme', isDarkMode.value ? 'dark' : 'light');
  };

  const initTheme = () => {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
      isDarkMode.value = true;
      document.getElementsByTagName('html')[0].classList.add('dark');
    } else {
      isDarkMode.value = false;
      document.getElementsByTagName('html')[0].classList.remove('dark');
    }
  };

  return {
    isDarkMode,
    toggleTheme,
    initTheme,
  };
}
