import koi from "@/utils/axios.ts";

export interface QuestionSearchResult {
  question: string;
  answer: string;
  type: string;
  is_ai: boolean;
  search_time: number;
  sources: string[];
}

export interface PackageItem {
  id: number;
  name: string;
  type: "time" | "count" | "time_count";
  durationSeconds?: number;
  totalCount: number;
  aiCount: number;
  priceCents: number;
  status: number;
  limitCount: number;
  createdAt: string;
}

export interface PackageInstance {
  id: number;
  packageId: number;
  packageName: string;
  packageType: string;
  startsAt: string;
  expiresAt?: string;
  remainingCount: number;
  remainingAiCount: number;
  status: number;
}

export interface OrderItem {
  id: number;
  orderNo: string;
  packageId: number;
  packageName: string;
  provider: string;
  amountCents: number;
  payableCents: number;
  refundedCents: number;
  status: string;
  expiresAt: string;
  paidAt?: string;
  createdAt: string;
}

export interface ApiKeyView {
  prefix: string;
  masked: string;
  createdAt: string;
  lastUsedAt?: string;
}

export const searchQuestion = (params: { q: string; type?: string; options?: string; package_id?: number }) =>
  koi.get<{ code: number; message: string; data: QuestionSearchResult }>("/api/v1/search", params);

export const listPackages = () => koi.get<{ code: number; message: string; data: PackageItem[] }>("/api/v1/packages");

export const listMyPackages = () =>
  koi.get<{ code: number; message: string; data: PackageInstance[] }>("/api/v1/packages/my");

export const createOrder = (data: { packageId: number; provider?: string }) =>
  koi.post<{ code: number; message: string; data: { order: OrderItem; paymentUrl: string } }>("/api/v1/orders", data);

export const listMyOrders = () => koi.get<{ code: number; message: string; data: OrderItem[] }>("/api/v1/orders/my");

export const getApiKey = () => koi.get<{ code: number; message: string; data: ApiKeyView }>("/api/v1/api-key");

export const createApiKey = () =>
  koi.post<{ code: number; message: string; data: { key: string; info: ApiKeyView } }>("/api/v1/api-key");

export const getOcsConfig = () => koi.get<any>("/api/ocs/config", { key: "" });
