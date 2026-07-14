import { RouteRecordRaw } from "vue-router";
import { HOME_URL, LOGIN_URL } from "@/config";
import Layout from "@/layouts/index.vue";

export const layoutRouter: RouteRecordRaw[] = [
  {
    path: LOGIN_URL,
    name: "login",
    component: () => import("@/views/login/index.vue"),
    meta: { title: "menu.login.auth" }
  }
];

export const staticRouter: RouteRecordRaw[] = [
  {
    path: "/",
    name: "layout",
    component: Layout,
    redirect: HOME_URL,
    meta: { title: "题库调用系统", isVisible: "0" },
    children: []
  },
  {
    path: "/payment",
    component: Layout,
    meta: { title: "支付结果", isVisible: "0" },
    children: [
      {
        path: "result",
        name: "paymentResult",
        component: () => import("@/views/tiku/payment-result/index.vue"),
        meta: { title: "支付结果", isVisible: "0" }
      }
    ]
  }
];

export const errorRouter: RouteRecordRaw[] = [
  { path: "/403", name: "403", component: () => import("@/views/error/403.vue") },
  { path: "/404", name: "404", component: () => import("@/views/error/404.vue") },
  { path: "/500", name: "500", component: () => import("@/views/error/500.vue") },
  { path: "/:pathMatch(.*)*", component: () => import("@/views/error/404.vue") }
];
