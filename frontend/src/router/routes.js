const routes = [
  // routes non protégées
  {
    path: "",
    component: () => import("src/layouts/MainLayout.vue"),
    meta: { requiresAuth: false },
    children: [
      {
        path: "",
        component: () => import("src/pages/HomePage.vue"),
      },
      {
        path: "appointments",
        component: () => import("src/pages/AppointmentPage.vue"),
      },
      {
        path: "validate-appointment",
        component: () => import("src/pages/AnonymousAppointmentPage.vue"),
        props: (route) => ({ token: route.query.token }),
      },
    ],
  },
  // routes protégées
  {
    path: "/",
    component: () => import("layouts/MainLayout.vue"),
    meta: { requiresAuth: true },
    children: [
      // /user
      {
        path: "user",
        component: () => import("src/layouts/EmptyLayout.vue"),
        children: [
          {
            path: "",
            component: () => import("src/pages/HomePage.vue"),
          },
          {
            path: "account",
            component: () => import("src/pages/AccountPage.vue"),
          },
        ],
      },
      // /director
      {
        path: "director",
        component: () => import("src/layouts/EmptyLayout.vue"),
        meta: { checkRole: true, requiredRole: 1 },
        children: [
          {
            path: "settings",
            component: () => import("src/layouts/EmptyLayout.vue"),
            children: [
              {
                path: "",
                component: () => import("src/pages/DirectorSettingsPage.vue"),
              },
            ],
          },
        ],
      },
      // /admin
      {
        path: "admin",
        component: () => import("src/layouts/EmptyLayout.vue"),
        meta: { checkRole: true, requiredRole: 0 },
        children: [
          {
            path: "appSettings",
            component: () => import("src/pages/AdminSettingsPage.vue"),
          },
        ],
      },
    ],
  },
  // routes d'erreur
  {
    path: "/:catchAll(.*)*",
    component: () => import("pages/ErrorNotFound.vue"),
  },
];

export default routes;
