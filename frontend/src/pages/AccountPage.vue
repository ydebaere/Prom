<template>
  <q-page class="q-pa-md">
    <!-- Informations de l'utilisateur -->
    <q-card
      class="q-mb-md"
      flat
      bordered
      style="max-width: 600px; margin: auto"
    >
      <q-card-section>
        <q-list dense>
          <q-item>
            <q-item-section side>
              <strong>Prénom :</strong>
            </q-item-section>
            <q-item-section>
              {{ user.given_name }}
            </q-item-section>
          </q-item>
          <q-item>
            <q-item-section side>
              <strong>Nom :</strong>
            </q-item-section>
            <q-item-section>
              {{ user.family_name }}
            </q-item-section>
          </q-item>
          <q-item>
            <q-item-section side>
              <strong>Email :</strong>
            </q-item-section>
            <q-item-section>
              {{ user.unique_name }}
            </q-item-section>
          </q-item>
          <q-item>
            <q-item-section side>
              <strong>Télécharger vos rendez-vous :</strong>
              <q-btn
                flat
                color="secondary"
                label="Télécharger"
                @click="downloadICS"
              />
            </q-item-section>
          </q-item>
          <!-- <q-item>
            <q-item-section side>
              <strong>Importer vos rendez-vous :</strong>
              <q-btn
                flat
                color="secondary"
                label="Importer"
                @click="uploadICS"
              />
            </q-item-section>
          </q-item> -->
        </q-list>
      </q-card-section>
      </q-card>
      <q-space class="q-mb-md" />
      <!-- Table des rendez-vous -->
      <q-card
        v-if="appointments.length > 0"
        flat
        bordered
        class="q-mt-md"
        style="max-width: 1000px; margin: auto"
      >
        <q-card-section>
          <div class="text-h5 q-mb-md">Vos rendez-vous</div>
            <q-table
            v-if="appointments.some(a => a.status === true)"
            flat
            bordered
            :rows="appointments.filter(a => a.status === true)"
            :columns="isMobile ? mobileColumns : columns"
            row-key="id"
            >
            <template v-if="isMobile" v-slot:body-cell-action="props">
              <q-td align="center">
                <q-btn
                  flat
                  round
                  dense
                  icon="visibility"
                  color="primary"
                  @click="viewDetails(props.row)"
                />
                <q-btn
                  v-if="props.row.guest === user.name"
                  flat
                  round
                  dense
                  icon="update"
                  color="primary"
                  @click="openEditDialog(props.row)"
                />
                <q-btn
                  flat
                  round
                  dense
                  icon="delete"
                  color="negative"
                  @click="deleteAppointments(props.row)"
                />
              </q-td>
            </template>

            <template v-else v-slot:body-cell-action="props">
              <q-td :props="props">
                <q-btn
                v-if="props.row.guest === user.name"
                  flat
                  round
                  dense
                  icon="edit"
                  color="secondary"
                  @click="openEditDialog(props.row)"
                />
                <q-btn
                  flat
                  round
                  dense
                  icon="delete"
                  color="negative"
                  @click="deleteAppointments(props.row)"
                />
              </q-td>
            </template>
          </q-table>
        </q-card-section>
      </q-card>
      <q-space class="q-mb-md" />
      <!-- Horaires de travail -->
      <q-card v-if="[0, 1, 2].some((role) => user.roles?.includes(role))"
        flat
        bordered
        class="q-mt-md"
        style="max-width: 1000px;
        margin: auto">
        <q-card-section>
          <q-btn
            label="Afficher/Masquer les horaires de travail"
            color="secondary"
            class="q-mt-md"
            @click="showWorkSchedule = !showWorkSchedule"
          />
          <div v-if="showWorkSchedule" class="column q-gutter-md">
            <q-space class="q-mb-md" />
            <div class="text-h5 q-mb-md">Vos horaires de travail</div>
            <div class="text-h8 q-mb-sm"> (n'oubliez pas d'enregistrer tout changements)</div>
            <div v-for="day in weekDays" :key="day" class="q-pa-sm" style="border: 1px solid #ccc; border-radius: 4px;">
              <div class="text-h6 q-mb-sm">{{ day }}</div>
              <!-- Liste des indispos -->
              <div v-for="(slot, index) in dailyAvailabilities[day]" :key="index" class="row q-col-gutter-md items-center q-mb-sm">
          <div class="col">
            <q-input v-model="slot.start" type="time" label="Début" dense outlined />
          </div>
          <div class="col">
            <q-input v-model="slot.end" type="time" label="Fin" dense outlined />
          </div>
          <div class="col-auto">
            <q-btn dense icon="delete" color="negative" flat @click="removeWorkSchedule(day, index)" />
          </div>
            </div>
            <!-- Ajouter une plage -->
            <q-btn
            icon="add"
            label="Ajouter une plage"
            dense
            flat
            color="secondary"
            @click="addAvailability(day)"
                />
              </div>
              <q-btn
                label="Enregistrer"
                color="secondary"
                class="q-mt-md"
                @click="saveWorkSchedule()"
              />
          </div>
        </q-card-section>
      </q-card>
      <q-space class="q-mb-md" />
        <!-- Indisponibilités Fixes -->
      <q-card v-if="[0, 1, 2].some((role) => user.roles?.includes(role))"
        flat
        bordered
        class="q-mt-md"
        style="max-width: 1000px;
        margin: auto">
        <q-card-section>
          <q-btn
            label="Afficher/Masquer les indisponibilités"
            color="secondary"
            class="q-mt-md"
            @click="showUnavailabilities = !showUnavailabilities"
          />
          <div v-if="showUnavailabilities" class="col-12 col-md-6">
            <q-space class="q-mb-md" />
            <div class="text-h5 q-mb-md">Indisponibilités</div>
            <q-btn
              flat
              color="secondary"
              icon="add"
              label="Ajouter une indisponibilité"
              @click="showDialog = true"
            />
            <q-dialog v-model="showDialog" persistent>
              <q-card style="min-width: 400px">
              <q-card-section>
                <div class="text-h6">Ajouter une indisponibilité</div>
              </q-card-section>
              <q-card-section>
                <q-form @submit.prevent="addUnavailability()">
                <q-input
                  v-model="newUnavailability.day"
                  label="Jour"
                  type="date"
                  outlined
                  dense
                  class="q-mb-sm">
                </q-input>
                <q-input
                  v-model="newUnavailability.start"
                  label="Début"
                  type="time"
                  outlined dense class="q-mb-sm"
                />
                <q-input
                  v-model="newUnavailability.end"
                  label="Fin"
                  type="time"
                  outlined dense class="q-mb-sm"
                />
                </q-form>
              </q-card-section>
              <q-card-actions align="right">
                <q-btn
                  flat
                  label="Annuler"
                  color="primary"
                  @click="showDialog = false"
                />
                <q-btn
                  flat
                  label="Ajouter"
                  color="secondary"
                    @click="addUnavailability(); showDialog = false"
                />
              </q-card-actions>
              </q-card>
            </q-dialog>
            <q-list bordered separator>
              <q-item v-for="(item, index) in unavailabilities" :key="index">
                <q-item-section>{{ item.reason }} le {{ item.date }} de {{ item.start }} à {{ item.end }}</q-item-section>
                <q-item-section side>
                    <q-btn v-if="item.reason != 'Rendez-vous'" flat icon="delete" color="negative" @click="removeUnavailability(item.id)" />
                </q-item-section>
              </q-item>
            </q-list>
          </div>
        </q-card-section>
      </q-card>
      <!-- dialog pour modifier un rendez-vous -->
      <q-dialog v-model="showEditDialog" persistent>
        <q-card style="min-width: 400px">
          <q-card-section>
            <div class="text-h6">Modifier le rendez-vous</div>
              </q-card-section>
                <q-card-section>
                  <q-input
                    v-model="editedAppointment.date"
                    label="Date"
                    type="date"
                    outlined
                    dense
                    class="q-mb-sm"
                    @update:model-value="loadAvailabilities({ ...editedAppointment})"
                    />
                    <q-select
                    v-model="editedAppointment.start"
                    :options="availabilities.length ? availabilities.map(slot => ({ label: `${slot.start}`, value: slot.start })) : []"
                    label="Plage horaire"
                    outlined
                    dense
                    class="q-mb-sm"
                    emit-value
                    map-options
                    />
                  <div v-if="noAvailabilitiesMessage" class="text-negative q-mt-sm">{{ noAvailabilitiesMessage }}</div>
                </q-card-section>
              <q-card-actions align="right">
            <q-btn flat label="Annuler" color="negative" v-close-popup />
            <q-btn flat label="Valider" color="secondary" @click="confirmUpdate(editedAppointment)" />
          </q-card-actions>
        </q-card>
      </q-dialog>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useQuasar } from "quasar";
import { fetchUserUnavailabilities, insertUnavailability, deleteUnavailability } from "src/services/unavailabilities";
import {
  deleteAppointment,
  fetchAppointmentsByUserID,
  updateAppointment,
  getAvailabilities,
} from "src/services/appointment";
import { getUser } from "src/services/api";
import {
  fetchWorkSchedules,
  createWorkSchedule,
  updateWorkSchedule,
  deleteWorkSchedule
} from "src/services/workSchedule";

defineOptions({
  name: "AccountPage",
});

const showWorkSchedule = ref(false);
const weekDays = [
  "Lundi",
  "Mardi",
  "Mercredi",
  "Jeudi",
  "Vendredi",
];
const dailyAvailabilities = ref({
  Lundi: [],
  Mardi: [],
  Mercredi: [],
  Jeudi: [],
  Vendredi: [],
  Samedi: [],
  Dimanche: [],
});
const showDialog = ref(false);
const showUnavailabilityDialog = ref(false);
const showUnavailabilities = ref(false);
const newUnavailability = ref({
  day: '',
  start: '',
  end: ''
});
const user = getUser();
const isMobile = ref(window.innerWidth <= 768);
const $q = useQuasar();
const appointments = ref([]);
const unavailabilities = ref([]);
const columns = [
  {
    name: "date",
    required: true,
    label: "Date",
    align: "left",
    field: (row) => row.start_time,
    format: (val) => formatDate(val, isMobile.value),
    sortable: true,
  },
  {
    name: "title",
    required: true,
    label: "Titre",
    align: "left",
    field: (row) => row.title,
    format: (val) => `${val}`,
    sortable: true,
  },
  {
    name: "Ecole",
    required: true,
    label: "Ecole",
    align: "left",
    field: (row) => row.school,
    format: (val) => `${val}`,
    sortable: true,
  },
  {
    name: "action",
    label: "Action",
    align: "center",
  },
];
const mobileColumns = [
  {
    name: "title",
    required: true,
    label: "Date",
    align: "left",
    field: (row) => row.start_time,
    format: (val) => formatDate(val, isMobile.value),
    sortable: true,
  },
  {
    name: "action",
    label: "Action",
    align: "center",
  },
];
const showEditDialog = ref(false);
const editedAppointment = ref({
  id: null,
  date: '',
  start: '',
  end: '',
});
const availabilities = ref([]);
const noAvailabilitiesMessage = ref("");

async function loadUnavailabilities() {
  try {
    const data = await fetchUserUnavailabilities(user.unique_name);
    unavailabilities.value = (data || [])
      .filter(item => item.reason.toLowerCase() !== 'rendez-vous')
      .map((item) => ({
        id: item.id,
        label: item.label,
        date: new Date(item.date).toLocaleDateString('fr-FR'),
        reason: item.reason,
        start: item.start_time.slice(11, 16),
        end: item.end_time.slice(11, 16),
      }));
  } catch (error) {
    console.error("Failed to fetch unavailabilities:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors du chargement des indisponibilités",
    });
  }
}

async function addUnavailability() {
  if (!newUnavailability.value.day || !newUnavailability.value.start || !newUnavailability.value.end) {
    $q.notify({
      type: "negative",
      message: "Veuillez remplir tous les champs",
    });
    return;
  }

  const payload = {
    user_id: user.unique_name,
    day: newUnavailability.value.day,
    start_time: newUnavailability.value.start,
    end_time: newUnavailability.value.end,
    reason: "Block"
  };

  try {
    await insertUnavailability(payload);
    $q.notify({
      type: "positive",
      message: "Indisponibilité ajoutée avec succès",
    });
    loadUnavailabilities();
    newUnavailability.value = { day: '', start: '', end: '' };
  } catch (error) {
    console.error("Failed to add unavailability:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de l'ajout de l'indisponibilité",
    });
  }
}

async function removeUnavailability(id) {
  try {
    await deleteUnavailability(id);
    $q.notify({
      type: "positive",
      message: "Indisponibilité supprimée avec succès",
    });
    loadUnavailabilities();
  } catch (error) {
    console.error("Failed to delete unavailability:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la suppression de l'indisponibilité",
    });
  }
}

async function loadAvailabilities(appointment) {
  const hostID = appointment.host;
  const schoolID = appointment.school;
  const date = appointment.start_time ? new Date(appointment.start_time) : appointment.date;
  const resource = appointment.resource;
  var duration = 0;

  if (resource == 1) {
    duration = 15;
  } else if (resource == 2) {
    duration = 30;
  } else {
    duration = 60;
  }

  try {
    const availabilitiesResponse = await getAvailabilities(
      hostID,
      date,
      resource
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

function openEditDialog(appointment) {
  console.log("Opening edit dialog for appointment:", appointment);
  loadAvailabilities(appointment);

  const start = new Date(appointment.start_time);
  const end = new Date(appointment.end_time);

  editedAppointment.value = {
    host: appointment.host,
    guest: appointment.guest,
    school: appointment.school,
    resource: appointment.resource,
    id: appointment.id,
    date: start.toISOString().slice(0, 10),
    start: null,
    end: null,
  };
  showEditDialog.value = true;
}

async function confirmUpdate(appointment) {
  const selectedSlot = availabilities.value.find(slot => slot.start === appointment.start);
  appointment.end = selectedSlot ? selectedSlot.end : null;

  console.log("Updating appointment:", appointment);
  console.log("Selected start time:", appointment.start);
  console.log("Selected end time:", appointment.end);
  try {
    await updateAppointment({
      appointment
    });

    $q.notify({
      type: "positive",
      message: "Rendez-vous modifié avec succès",
    });

    showEditDialog.value = false;
    loadUserAppointments();
  } catch (error) {
    $q.notify({
      type: "negative",
      message: "Erreur lors de la modification",
    });
  }
}

function addAvailability(day) {
  dailyAvailabilities.value[day].push({
  });
}

async function loadUserAppointments() {
  const userID = user.unique_name;
  try {
    const data = await fetchAppointmentsByUserID(userID);
    appointments.value = data || [];
  } catch (error) {
    console.error("Failed to fetch user appointments:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors du chargement des rendez-vous",
    });
  }
}

async function loadWorkSchedules() {
  try {
    const data = await fetchWorkSchedules(user.unique_name);

    const grouped = {
      Lundi: [],
      Mardi: [],
      Mercredi: [],
      Jeudi: [],
      Vendredi: [],
      Samedi: [],
      Dimanche: [],
    };

    Object.keys(data).forEach((day) => {
      const dayCapitalized = capitalizeFirst(day);
      data[day].forEach((item) => {
        grouped[dayCapitalized].push({
          id: item.id,
          start: item.start_time.slice(11, 16),
          end: item.end_time.slice(11, 16),
        });
      });
    });

    dailyAvailabilities.value = grouped;
  } catch (error) {
    $q.notify({
      type: "negative",
      message: "Erreur lors du chargement des horaires de travail",
    });
  }
}

async function saveWorkSchedule() {
  try {
    const userId = user.unique_name;
    for (const day in dailyAvailabilities.value) {
      for (const slot of dailyAvailabilities.value[day]) {
        if (!slot.start || !slot.end) {
          $q.notify({
            type: "negative",
            message: `Veuillez remplir tous les champs pour ${day}`,
          });
          return;
        }
        if (slot.start >= slot.end) {
          $q.notify({
            type: "negative",
            message: `L'heure de début doit être avant l'heure de fin pour ${day}`,
          });
          return;
        }
        const payload = {
          user_id: userId,
          day_of_week: day.toLowerCase(),
          start_time: slot.start,
          end_time: slot.end,
        };

        if (slot.id) {
          await updateWorkSchedule({ ...payload, id: slot.id });
        } else {
          await createWorkSchedule(payload);
        }
      }
    }
    $q.notify({
      type: "positive",
      message: "Horaires enregistrés avec succès",
    });
    loadWorkSchedules();
  } catch (error) {
    $q.notify({
      type: "negative",
      message: "Erreur lors de l'enregistrement des horaires",
    });
  }
}

async function removeWorkSchedule(day, index) {
  const slot = dailyAvailabilities.value[day][index];
  $q.dialog({
    title: "Confirmation",
    message: "Êtes-vous sûr de vouloir supprimer cet horaire de travail ?",
    ok: {
      label: "Oui",
      color: "negative",
    },
    cancel: {
      label: "Non",
      color: "primary",
    },
  }).onOk(async () => {
    if (slot.id) {
      await deleteWorkSchedule(slot.id, user.unique_name);
    }
    dailyAvailabilities.value[day].splice(index, 1);
    $q.notify({
      type: "positive",
      message: "Horaire supprimé avec succès",
      position: "center",
    });
  }).onCancel(() => {
    $q.notify({
      type: "info",
      message: "Suppression annulée",
      position: "center",
    });
  });
}

function capitalizeFirst(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function downloadICS() {
  const link = document.createElement("a");
  const server = import.meta.env.VITE_BASE_URL;
  link.href = `${server}calendar-ics?unique_name=${user.unique_name}`;
  link.download = "mon_calendrier.ics";
  link.click();
}

function formatDate(date, isMobile) {
  const adjustedDate = new Date(date);
  adjustedDate.setHours(adjustedDate.getHours() - 2);
  const options = isMobile
    ? { year: "numeric", month: "short", day: "numeric" }
    : {
        year: "numeric",
        month: "long",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      };
  return adjustedDate.toLocaleString("fr-FR", options);
}

async function deleteAppointments(appointment) {
  try {
    await deleteAppointment(appointment.id);
    $q.notify({
      type: "positive",
      message: "Rendez-vous supprimé avec succès",
    });
    loadUserAppointments();
  } catch (error) {
    console.error("Failed to delete appointment:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la suppression du rendez-vous",
    });
  }
}

async function updateAppointments(appointment) {
  console.log("Updating appointment:", appointment);
  const updatedData = {
    id: appointment.id,
    date: appointment.date,
    start_time: appointment.start_time,
    end_time: appointment.end_time,
  };
  try {
    await updateAppointment(updatedData);
    $q.notify({
      type: "positive",
      message: "Rendez-vous modifié avec succès",
    });
    loadUserAppointments();
  } catch (error) {
    console.error("Failed to update appointment:", error);
    $q.notify({
      type: "negative",
      message: "Erreur lors de la modification du rendez-vous",
    });
  }
}

function viewDetails(row) {
  $q.dialog({
    title: "Détails du rendez-vous",
    message: `
      <div><strong>Date:</strong> ${row.start_time}</div>
      <div><strong>Titre:</strong> ${row.title}</div>
      <div><strong>Ecole:</strong> ${row.school}</div>
    `,
    html: true,
    ok: {
      label: "Fermer",
      color: "secondary",
    },
  });
}

onMounted(() => {
  loadUserAppointments();
  loadWorkSchedules();
  loadUnavailabilities();
});
</script>

<style lang="scss">
.my-class {
  background-color: $primary;
  color: $policy;
}
</style>
