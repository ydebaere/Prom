<template>
  <q-layout
    view="lHh Lpr lff"
    container
    style="
      height: 100vh;
      width: 100vw;
      max-height: 100vh;
      max-width: 100vw;
      margin: auto;
    "
    class="rounded-borders no-border"
  >
    <q-header elevated class="my-class">
      <q-toolbar>
        <q-toolbar-title @click="navigateTo('/')" class="cursor-pointer"
          >Prometheus</q-toolbar-title
        >
        <q-btn
          flat
          v-if="user"
          label="Mon compte"
          @click="navigateTo('/user/account')"
          icon="settings"
          class="account-btn"
        />
        <q-btn
          flat
          v-if="user != null && user.roles?.includes(0) && !isMobile"
          label="Administration"
          @click="navigateTo('/admin/appSettings')"
          icon="settings"
          class="account-btn"
        />
        <q-btn
          flat
          v-if="user != null && user.roles?.includes(1) && !isMobile"
          label="Mon établissement"
          @click="navigateTo('/director/settings')"
          icon="settings"
          class="account-btn"
        />
        <q-btn
          v-if="user"
          flat
          @click="logout"
          icon="logout"
          class="logout-btn"
        />
        <q-btn
          flat
          v-if="!user"
          label="Se connecter avec Azure"
          icon="login"
          @click="loginWithAzure"
          class="account-btn"
        />
      </q-toolbar>
    </q-header>
    <q-footer class="my-class footer">
      <q-toolbar>
        <q-toolbar-title @click="navigateTo('/')" class="cursor-pointer"
          >Prometheus</q-toolbar-title
        >
        <q-btn
          flat
          label="Contact"
          href="mailto:ydebaere@gmail.com?subject=Demande de contact Prometheus"
          class="account-btn"
        />
        <q-btn
          flat
          label="Mentions légales"
          @click="navigateTo('/legal')"
          class="account-btn"
        />
        <q-btn
          flat
          label="Conditions d'utilisation"
          @click="navigateTo('/terms')"
          class="account-btn"
        />
        <q-btn
          flat
          label="Politique de confidentialité"
          @click="navigateTo('/privacy')"
          class="account-btn"
        />
      </q-toolbar>
    </q-footer>
    <q-page-container class="layout-container">
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { getUser } from "src/services/api.js";
import * as msal from "@azure/msal-browser";

export default {
  setup() {
    const $q = useQuasar();
    const drawer = ref(false);
    const router = useRouter();
    const user = getUser();
    const email = ref("");
    const password = ref("");
    const passwordInput = ref(null);
    const isMobile = ref(window.innerWidth <= 768);
    const showLoginDialog = ref(false);
    const privacyPolicy = ref(false);
    const isMsalReady = ref(false);
    const isPwd = ref(true);

    const msalInstance = new msal.PublicClientApplication({
      auth: {
        clientId: "b134585a-b99c-485a-92d7-0f331904667d",
        authority:
          "https://login.microsoftonline.com/cde3eff2-d1ad-46a3-9340-d7a530d15963",
        redirectUri: "/",
      },
    });

    onMounted(async () => {
      try {
        await msalInstance.initialize();
        isMsalReady.value = true;
      } catch (error) {
        console.error("Erreur MSAL init :", error);
      }
      window.addEventListener("resize", () => {
        isMobile.value = window.innerWidth <= 768;
      });
    });

    function logout() {
      localStorage.clear();
      router.push("/");
      window.location.reload();
    }

    function navigateTo(path) {
      router.push(path);
    }

    async function loginWithAzure() {
      if (!isMsalReady.value) {
        $q.notify({
          type: "warning",
          message: "MSAL n'est pas encore prêt. Patientez...",
        });
        return;
      }
      try {
        await msalInstance.loginRedirect({
          scopes: ["api://a09c54f3-0c53-49d4-b9ef-16eaecdd5b74/Access"],
        });
      } catch (error) {
        console.error("Azure Login Redirect Error:", error);
        $q.notify({
          type: "negative",
          message: "Erreur lors de la redirection pour la connexion Azure",
        });
      }
    }

    return {
      loginWithAzure,
      drawer,
      user,
      logout,
      navigateTo,
      isMobile,
      showLoginDialog,
      email,
      password,
      privacyPolicy,
      passwordInput,
      isPwd,
      isMsalReady,
    };
  },
};
</script>

<style lang="scss" scoped>
.my-class {
  background-color: $secondary;
  color: $policy;
}

.my-class2 {
  background-color: $accent;
  color: $policy;
}

.footer {
  margin-top: auto;
}

.layout-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.no-underline {
  text-decoration: none;
}
.full-width-btn {
  width: 95%;
  margin-bottom: 10px;
}
.column-buttons {
  display: flex;
  flex-direction: column;
}
.full-width-btn {
  width: 100%;
  margin-bottom: 10px;
}
.account-btn {
  margin-right: 10px;
}
.logout-btn {
  margin-left: 10px;
}
</style>
