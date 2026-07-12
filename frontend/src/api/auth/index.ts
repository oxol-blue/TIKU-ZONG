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

export const login = (data: { email: string; password: string }) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/login", data);

export const register = (data: { email: string; password: string }) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/register", data);

export const refresh = (refreshToken: string) =>
  koi.post<{ code: number; message: string; data: AuthSession }>("/api/v1/auth/refresh", { refreshToken });

