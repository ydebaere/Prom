<template>
  <q-page class="q-pa-md">
    <q-card class="q-mb-md no-border">
      <q-card-section class="flex flex-center">
        <q-btn
          class="my-class"
          label="Prendre rendez-vous!"
          @click="navigateTo('/appointments')"
        />
      </q-card-section>
    </q-card>

    <!-- Prochains rendez-vous -->
    <q-card class="q-mb-md" v-if="nextAppointments.length > 0">
      <q-card-section class="my-class">
        <div class="row q-col-gutter-md justify-start">
          <div
            v-for="(appointment, index) in nextAppointments"
            :key="index"
            class="col-xs-12 col-sm-6 col-md-4"
          >
            <q-card flat bordered class="my-class q-pa-md">
              <div class="q-mb-sm">
                Le
                <strong>{{ formatDate(appointment.start_time) }}</strong> à
                <strong>{{ formatTime(appointment.start_time) }} </strong>
              </div>
              <div class="q-mb-sm">
                Rendez-vous avec
                <strong>{{
                  appointment.host === user.unique_name
                    ? appointment.guest
                    : appointment.host
                }}</strong>
              </div>
              <div class="q-mt-sm">
                Lieux : <strong>{{ appointment.school }}</strong>
              </div>
              <div class="q-mt-sm">
                Durée :
                <strong>{{
                  calculateDuration(
                    appointment.start_time,
                    appointment.end_time
                  )
                }}</strong>
                mins
              </div>
            </q-card>
          </div>
        </div>
      </q-card-section>
    </q-card>

    <q-card class="q-mb-md" flat>
      <q-card-section class="my-class text-center">
        <div class="text-h6">Nos Écoles partenaires</div>
      </q-card-section>
    </q-card>

    <q-space class="q-mb-md" style="height: 20px" />

    <q-card-section class="row q-col-gutter-md justify-around">
      <q-card
        flat
        bordered
        clickable
        style="width: 30%"
        @click="openExternalLink1"
      >
        <q-img src="../assets/LOGO_HLF.png" contain />
      </q-card>

      <q-card
        flat
        bordered
        clickable
        style="width: 30%"
        @click="openExternalLink2"
      >
        <q-img src="../assets/LOGO_LT.png" contain />
      </q-card>
    </q-card-section>

    <q-space class="q-mb-md" style="height: 20px" />

    <q-card class="q-mb-md" flat v-if="user?.unique_name === 'steve.colin@hainaut-promsoc.be'">
      <q-card-section class="my-class text-center">
      <div class="text-h6">@STEVE</div>
      <q-img
        src="https://media2.giphy.com/media/v1.Y2lkPTc5MGI3NjExNmtsNHp4NmtzbW1yeDc4eGdpOWowNGF1MmIzaDVtZ2phNmR6bHhmciZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/U6vnlPWryhe64Ac2rp/giphy.gif"
        style="max-width: 300px; margin: 0 auto;"
        contain
      />
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { fetchAppointmentsByUserID } from "src/services/appointment";
import { checkTokenExpiry, getUser } from "src/services/api";

defineOptions({
  name: "homePage",
});

const user = getUser();
const router = useRouter();
const nextAppointments = ref([]);
const date = ref(null);

function openExternalLink1() {
  window.open(
    "https://www.etudierenhainaut.be/institut-provincial-henri-la-fontaine/",
    "_blank"
  );
}

function openExternalLink2() {
  window.open(
    "https://www.etudierenhainaut.be/lise-thiry/",
    "_blank"
  );
}

function navigateTo(path) {
  router.push(path);
}

function formatTime(time) {
  return time ? time.slice(11, 16) : "";
}

function formatDate(date) {
  if (!date) return "";

  const dateObj = new Date(date);
  const day = String(dateObj.getDate()).padStart(2, "0");
  const month = String(dateObj.getMonth() + 1).padStart(2, "0");
  const year = dateObj.getFullYear();

  return `${day}/${month}/${year}`;
}

function calculateDuration(startTime, endTime) {
  const start = new Date(startTime);
  const end = new Date(endTime);
  const duration = (end - start) / (1000 * 60);
  return duration;
}

async function loadNextAppointment() {
  if (!localStorage.getItem("user")) return;
  const user = JSON.parse(localStorage.getItem("user"));
  const userID = user.unique_name;
  try {
    const data = await fetchAppointmentsByUserID(userID);
    if (!data || data.length === 0) {
      nextAppointments.value = [];
      return;
    }
    console.log("Appointments data:", data);
    const now = new Date();
    const upcomingAppointments = data.filter(
      (appointment) => new Date(appointment.start_time) > now && appointment.status === true
    );
    nextAppointments.value = upcomingAppointments.slice(0, 3);
  } catch (error) {
    console.error("Error fetching appointments:", error.message || error);
  }
}

function setDate() {
  const now = new Date();
  date.value = formatDate(now);
}

onMounted(() => {
  setInterval(checkTokenExpiry, 60000);
  setDate();
  loadNextAppointment();
  calculateDuration();
});
</script>

<style lang="scss" scoped>
.my-class {
  background-color: $secondary;
  color: $policy;
}
</style>
