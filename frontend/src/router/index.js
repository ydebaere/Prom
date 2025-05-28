import { route } from "quasar/wrappers";
import { createRouter, createWebHistory } from "vue-router";
import routes from "./routes";
import { Notify } from "quasar";
import { getUser } from "src/services/api";

export default route(function (/* { store, ssrContext } */) {
  const createHistory = createWebHistory;

  const Router = createRouter({
    scrollBehavior: () => ({ left: 0, top: 0 }),
    routes,

    history: createHistory(process.env.VUE_ROUTER_BASE),
  });

  // Middleware global pour la validation du token
  Router.beforeEach((to, from, next) => {
    const token = localStorage.getItem("token");
    const user = getUser();

    // Vérification des routes nécessitant une authentification
    if (to.matched.some((record) => record.meta.requiresAuth)) {
      if (!token || !user) {
        Notify.create({
          type: "warning",
          message: "Accès restreint : Vous  n'êtes pas connecté.",
        });
        return next({ path: "/" });
      }
    }

    // Redirection si le rôle utilisateur est >= 3
    if (to.matched.some((record) => record.meta.checkRole)) {
      const userRole = user?.roles?.[0];
      const requiredRole = to.meta.requiredRole || 0;
      if (userRole > requiredRole) {
        Notify.create({
          type: "warning",
          message:
            "Accès restreint : Votre rôle ne permet pas d'accéder à cette page.",
        });
        return next({ path: "/" });
      }
    }

    next();
  });

  return Router;
});
