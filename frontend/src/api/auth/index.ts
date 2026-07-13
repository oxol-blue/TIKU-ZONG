import koi from "@/utils/axios.ts";

export interface AuthUser {
  id: number;
  email: string;
  role: string;
  status: number;
  createdAt: string;
}

export interface AuthSession {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export const login = (data: { email: string; password: string; captchaId?: string; captchaCode?: string }) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/login", data);

export const getCaptcha = () =>
  koi.get<{ code: number; message: string; data: { captchaId: string; image: string } }>("/api/v1/auth/captcha");

export const register = (data: { email: string; password: string; inviteCode?: string }) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/register", data);

export const refresh = (refreshToken: string) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/refresh", { refreshToken });

export const getMe = () => koi.get<{ code: number; message: string; data: AuthUser }>("/api/v1/me");
export const changePassword = (data: { currentPassword: string; newPassword: string }) =>
  koi.post<{ code: number; message: string }>("/api/v1/password/change", data);
