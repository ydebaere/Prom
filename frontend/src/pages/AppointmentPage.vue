<template>
  <q-page :class="{ 'desktop-view': !isMobile }" class="q-pa-md q-mx-auto">
    <q-card-section>
      <!-- Step 1: Sélectionnez une école -->
      <q-select
        outlined
        v-model="model1"
        label="Etape 1: Sélectionnez une école"
        :dense="dense"
        :options-dense="denseOpts"
        :options="schools"
        option-label="name"
        option-value="id"
        placeholder="Sélectionnez une école"
      >
        <template v-slot:prepend>
          <q-icon name="home" />
        </template>
      </q-select>
    </q-card-section>

    <!-- Sélection du type de rendez-vous -->
    <q-card-section v-if="model1 && userData != null" class="q-mb-md">
      <!-- Step 2: Sélectionnez un type de rendez-vous -->
      <q-select
        outlined
        v-model="model2"
        label="Etape 2 : Sélectionnez un type de rendez-vous"
        :dense="dense"
        :options-dense="denseOpts"
        :options="resources.filter((resource) => resource.duration > 0)"
        option-label="description"
        @focus="loadResources"
      >
        <template v-slot:prepend>
          <q-icon name="work" />
        </template>
      </q-select>
    </q-card-section>
    <div
      v-if="model2"
      class="q-mb-md flex flex-center text-grey-7"
    >
      Etape 3 : Selectionnez une date
    </div>
    <!-- Calendrier et disponibilités -->
    <q-card-section
      v-if="model2 || (model1 && userData == null)"
      flat
      class="q-mb-md flex flex-center"
    >
      <!-- Step 3: Sélectionnez une date -->
      <q-date v-model="date" :options="dateOptions" color="secondary" />
    </q-card-section>

    <!-- Sélection de la ressource et disponibilités -->
    <q-card-section :class="isMobile ? '' : 'row items-start'">
      <!-- Ressource -->
      <q-select
        v-if="date && userData"
        outlined
        v-model="model4"
        label="Etape 4 : Sélectionnez un agent"
        :options="ressources"
        :dense="dense"
        :options-dense="denseOpts"
        :option-label="formatLabel"
        class="col-12 col-md-6"
        @focus="loadAgents"
      >
        <template v-slot:no-option>
          <div class="text-grey text-center q-pa-sm">
            Pas d'agent disponible pour cette ressource
          </div>
        </template>
        <template v-slot:prepend>
          <q-icon name="people" />
        </template>
      </q-select>

      <!-- Disponibilités -->
      <q-select
        v-if="model4"
        outlined
        v-model="model5"
        :options="availabilities"
        :dense="dense"
        :options-dense="denseOpts"
        option-label="start"
        option-value="end"
        label="Etape 5 : Séléctionnez une disponibilité"
        class="col-12 col-md-6"
        @focus="loadAvailabilities"
      >
        <template v-slot:prepend>
          <q-icon name="event" />
        </template>
      </q-select>

      <q-banner
        v-if="noAvailabilitiesMessage"
        class="bg-grey-2 text-negative q-my-sm col-12"
        dense
        rounded
      >
        {{ noAvailabilitiesMessage }}
      </q-banner>

      <!-- etape intermédiare pour invité-->
      <q-select v-if="userData == null && date"
        outlined
        v-model="timing"
        :options="timings"
        :dense="dense"
        :options-dense="denseOpts"
        option-label="start"
        option-value="end"
        label="Etape 5 : Séléctionnez une heure souhaitée"
        class="col-12 col-md-6"
        @focus="loadAvailabilities"
      >
      </q-select>

      <!-- Note -->
      <q-input
        v-if="model5 || (userData == null && date)"
        v-model="model6"
        label="Etape 6: Ajoutez une note (optionnel)"
        outlined
        dense
        clearable
        type="textarea"
        class="col-12"
      />

      <!-- Bouton -->
      <q-btn
        v-if="model5 || (userData == null && date)"
        class="q-mt-sm col-12"
        label="Planifier votre rendez-vous"
        color="secondary"
        @click="submitAppointment"
      />
    </q-card-section>

    <!-- Modal Email Invité -->
    <q-dialog v-model="showEmailDialog">
      <q-card>
        <q-card-section>
          <div class="text-h6">Veuillez entrer votre adresse mail</div>
        </q-card-section>
        <q-card-section>
          <q-input
            v-model="guestEmail"
            label="Email"
            type="email"
            outlined
            dense
            clearable
          />
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Annuler" color="negative" v-close-popup />
          <q-btn
            flat
            label="Confirmer"
            color="secondary"
            @click="confirmEmail"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup>
import { ref, watch, onMounted } from "vue";
import { useQuasar } from "quasar";
import { fetchSchools } from "src/services/school";
import { fetchAgents as fetchAgents } from "src/services/user";
import { fetchResources } from "src/services/resource";
import { getAvailabilities, insertAppointment, getSchedules } from "src/services/appointment";
import { useRoute, useRouter } from "vue-router";
import { getUser } from "src/services/api";

const route = useRouter();
const currentRoute = useRoute();
const $q = useQuasar();
const model1 = ref(null);
const model2 = ref(null);
const model4 = ref(null);
const model5 = ref(null);
const availabilities = ref([]);
const timing = ref(null);
const timings = [
  "09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
  "12:00", "13:00", "13:30", "14:00", "14:30", "15:00",
  "15:30", "16:00", "17:00", "17:30", "18:00", "18:30",
"19:00", "19:30", "20:00", "20:30"];
const model6 = ref(null);
const date = ref(null);
const schools = ref([]);
const resources = ref([]);
const ressources = ref([]);
const noAvailabilitiesMessage = ref("");
const dense = ref(false);
const denseOpts = ref(false);
const isMobile = ref(window.innerWidth <= 768);
const userData = getUser();
const showEmailDialog = ref(false);
const guestEmail = ref("");

window.addEventListener("resize", () => {
  isMobile.value = window.innerWidth <= 768;
});

function dateOptions(date) {
  const day = new Date(date).getDay();
  return day !== 0 && day !== 6;
}

onMounted(async () => {
  const schoolIDFromUrl = currentRoute.query.schoolID;
  const typeIDFromUrl = currentRoute.query.typeID;
  const resourceIDFromUrl = currentRoute.query.resourceID;

  try {
    if (schools.value.length === 0) {
      await loadSchools();
    }

    // Pré-sélectionner l'école si `schoolID` est présent
    if (schoolIDFromUrl) {
      const school = schools.value.find(
        (school) => school.id == schoolIDFromUrl
      );
      if (school) {
        model1.value = school;
      }
    }

    // Charger les types si `typeID` est présent
    if (typeIDFromUrl) {
      if (resources.value.length === 0) {
        await loadResources();
      }
      // Pré-sélectionner le type si `typeID` est présent
      const type = resources.value.find((type) => type.id == typeIDFromUrl);
      if (type) {
        model2.value = type;
      }
    }

    // Charger les ressources si `resourceID` est présent
    if (resourceIDFromUrl) {
      if (ressources.value.length === 0) {
        await loadAgents();
      }
      // Pré-sélectionner la ressource si `resourceID` est présent
      const resource = ressources.value.find(
        (resource) => resource.id == resourceIDFromUrl
      );
      if (resource) {
        model4.value = resource;
      }
    }
  } catch (error) {
    console.error("Erreur lors du traitement des paramètres de l'URL :", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors du chargement des données",
    });
  }
});

async function loadSchools() {
  try {
    const data = await fetchSchools();
    schools.value = data || [];

    const schoolIDFromUrl = currentRoute.query.schoolID;
    if (schoolIDFromUrl) {
      const school = schools.value.find(
        (school) => school.id == schoolIDFromUrl
      );
      if (school) {
        model1.value = school;
      }
      const typeIDFromUrl = currentRoute.query.typeID;
      if (typeIDFromUrl) {
        const type = resources.value.find((type) => type.id == typeIDFromUrl);
        if (type) {
          model2.value = type;
        }
      }
      const resourceIDFromUrl = currentRoute.query.resourceID;
      if (resourceIDFromUrl) {
        const resource = ressources.value.find(
          (resource) => resource.id == resourceIDFromUrl
        );
        if (resource) {
          model4.value = resource;
        }
      }
    }
  } catch (error) {
    console.error("Failed to fetch schools:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la récupération des écoles",
    });
  }
}

async function loadResources() {
  try {
    const data = await fetchResources(model1.value.id);
    resources.value = data || [];
  } catch (error) {
    console.error("Failed to fetch resources:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la récupération des resources",
    });
  }
}

async function loadAgents() {
  const resourceID = model2.value.id;
  const schoolID = model1.value.id;
  try {
    const data = await fetchAgents(schoolID, resourceID);
    ressources.value = data || [];
  } catch (error) {
    console.error("Failed to fetch agents:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la récupération des agents",
    });
  }
}

function formatLabel(option) {
  return `${option.given_name} ${option.family_name}`;
}

async function loadAvailabilities() {
  try {
    if (!model4.value) {
      console.log("No resource selected");
    //   const slotsResponse = await getSchedules(
    //   model1.value.id,
    //   date.value,
    //   );
    //   availabilities.value = (slotsResponse || []).map((slot) => ({
    //   start: slot.start,
    //   end: slot.end,
    //   label: slot.label,
    // }));
    // noAvailabilitiesMessage.value = availabilities.value.length
    //   ? ""
    //   : "Pas de disponibilités";
    return;
    }
    console.log("Model 2:", model2.value);
    const duration = model2.value.id;

    const availabilitiesResponse = await getAvailabilities(
      model4.value.user_id,
      date.value,
      duration
    );

    availabilities.value = (availabilitiesResponse || []).map((slot) => ({
      start: slot.start,
      end: slot.end,
      label: slot.label,
    }));

    noAvailabilitiesMessage.value = availabilities.value.length
      ? ""
      : "Pas de disponibilités";
  } catch (error) {
    console.error("Failed to fetch availabilities:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la récupération des disponibilités",
    });
  }
}

async function submitAppointment() {
  if (userData == null) {
    showEmailDialog.value = true;
    return;
  }

  const user = JSON.parse(localStorage.getItem("user"));
  let title = `Rendez-vous avec ${
    model4.value.given_name + " " + model4.value.family_name
  }`;

  const formattedDate = date.value.replace(/\//g, "-");

  // Création de la date de début
  const startDate = new Date(`${formattedDate}T${model5.value.start}:00`);
  const startTimestamp = Math.floor(startDate.getTime() / 1000);

  // Création de la date de fin
  const endDate = new Date(startDate);
  endDate.setMinutes(endDate.getMinutes() + model2.value.duration);
  const endTimestamp = Math.floor(endDate.getTime() / 1000);

  const newAppointment = {
    school: model1.value.id,
    resource: model2.value.id,
    date: formattedDate,
    host: model4.value.user_id,
    guest: userData.unique_name,
    start_time: startTimestamp,
    end_time: endTimestamp,
    title: title,
  };

  const newDate = new Date();
  const now = new Date();
  const selectedDateTime = new Date(`${date.value.replace(/\//g, "-")}T${model5.value.start}:00`);
  if (selectedDateTime.getTime() < now.getTime() + 2 * 60 * 60 * 1000) {
    $q.notify({
      type: "negative",
      message: "Veuillez sélectionner une date ultérieure",
      position: "center",
    });
    return;
  }

  try {
    await insertAppointment(newAppointment);
    model1.value = null;
    model2.value = null;
    model4.value = null;
    model5.value = null;
    date.value = null;

    $q.notify({
      type: "positive",
      message: `Rendez-vous ajouté avec succès. Veuillez le confirmer via l'email reçu!`,
      position: "center",
    });
  } catch (error) {
    if (error.message === "unconfirmed") {
      $q.notify({
        type: "negative",
        message: "Vous avez déjà un rendez-vous non confirmé. Veuillez vérifier votre email.",
        position: "center",
      });
      route.push("/");
    }
    else {
      $q.notify({
      type: "negative",
      message: "Erreur lors de l'ajout du rendez-vous. Veuillez réessayer.",
      position: "center",
    });
    }
  }
}

async function confirmEmail() {
  if (!guestEmail.value) {
    $q.notify({
      type: "negative",
      message: "Veuillez entrer une adresse mail valide",
      position: "center",
    });
    return;
  }

  const newDate = new Date(Date.now());
  // Vérifie que la date/heure choisie est au moins 2 heures après maintenant
  const selectedDateTime = new Date(`${date.value.replace(/\//g, "-")}T${timing.value}:00`);
  if (selectedDateTime.getTime() < newDate.getTime() + 2 * 60 * 60 * 1000) {
    $q.notify({
      type: "negative",
      message: "Veuillez sélectionner une date ultérieure",
      position: "center",
    });
    return;
  } try {
    // Calculer startTime à partir de la date choisie et de l'heure sélectionnée (timing)
    const formattedDate = date.value.replace(/\//g, "-");
    const startTime = `${formattedDate} ${timing.value}:00`;

    const anonymousAppointment = {
      unique_name: guestEmail.value,
      school_id: model1.value.id,
      start_time: startTime,
      end_time: startTime,
      host: 2,
    };

    console.log("Anonymous Appointment Data:", anonymousAppointment);

    const server = import.meta.env.VITE_BASE_URL;
    const response = await fetch(
      `${server}anonymous-appointment`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(anonymousAppointment),
      }
    );
    if (response.status === 422) {
      $q.dialog({
      title: "Aucune disponibilité",
      message: "Pas de disponibilités pour cette heure/date, veuillez choisir une autre heure/date.",
      ok: "Fermer"
      });
      timing.value = null;
      date.value = null;
      showEmailDialog.value = false;
      return;
    }
    else if (response.status === 409) {
      $q.dialog({
      title: "Rendez-vous déjà existant",
      message: "Vous avez déjà un rendez-vous non confirmé. Veuillez vérifier votre email.",
      ok: "Fermer"
      });
      timing.value = null;
      date.value = null;
      showEmailDialog.value = false;
      return;
    }
    if (response.status >= 200 && response.status < 300) {
      $q.notify({
        type: "positive",
        message: `Rendez-vous ajouté avec succès n'oubliez pas de le confirmer via l'email recu!`,
        position: "center",
      });
      showEmailDialog.value = false;
      route.push("/");
    } else if (response.status === 500) {
      $q.notify({
        type: "negative",
        message: "Erreur serveur lors de l'ajout du rendez-vous anonyme",
        position: "center",
      });
      showEmailDialog.value = false;
    }
  } catch (error) {
    console.error("Erreur lors de la création du rendez-vous anonyme:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de l'ajout du rendez-vous anonyme",
      position: "center",
    });
  }
}

watch(model1, (newValue) => {
  model2.value = null;
  model4.value = null;
  model5.value = null;
  date.value = null;
  resources.value = [];
  ressources.value = [];
  availabilities.value = [];
});

watch(model2, (newValue) => {
  model4.value = null;
  model5.value = null;
  date.value = null;
  ressources.value = [];
  availabilities.value = [];
});

watch(date, (newValue) => {
  model4.value = null;
  model5.value = null;
  availabilities.value = [];
  noAvailabilitiesMessage.value = "";
});

watch(model4, (newValue) => {
  model5.value = null;
  availabilities.value = [];
  noAvailabilitiesMessage.value = "";
});
</script>

<style lang="scss" scoped>
.my-class {
  background-color: $primary;
  color: $policy;
}
.desktop-view {
  min-width: 800px;
}
</style>
