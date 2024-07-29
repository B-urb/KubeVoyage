import { writable } from 'svelte/store';

function createAuthStore() {
  const { subscribe, set } = writable(false);

  return {
    subscribe,
    setAuth: (value) => {
      set(value);
      if (typeof window !== 'undefined') {
        localStorage.setItem('isAuthenticated', JSON.stringify(value));
      }
    },
    checkAuth: () => {
      if (typeof window !== 'undefined') {
        const storedAuth = localStorage.getItem('isAuthenticated');
        if (storedAuth) {
          set(JSON.parse(storedAuth));
        }
      }
    }
  };
}

export const isAuthenticated = createAuthStore();
