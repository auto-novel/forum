import { createAuthApi, type AuthUser } from '@novelia/auth-api';
import { readonly, ref } from 'vue';

const authUrl = new URL(__AUTH_URL__, window.location.origin);

export const authApi = createAuthApi({
  app: 'f',
  baseUrl: new URL('api/v1/', authUrl).toString(),
  storage: {
    key: 'f-session',
    target: localStorage,
  },
});

const user = ref<AuthUser>();

authApi.subscribeUser((profile) => {
  user.value = profile;
});

export const authUser = readonly(user);
export const loginUrl = authUrl.toString();

void authApi.auth.refresh().catch(() => undefined);
