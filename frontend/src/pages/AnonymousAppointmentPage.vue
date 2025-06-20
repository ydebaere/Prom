<template>
  <q-layout view="lHh Lpr lFf">
    <q-page-container>
      <q-page class="q-pa-md">
        <div v-if="loading" class="q-mt-md">
          <q-spinner color="primary" size="50px" />
        </div>
        <div v-else>
          <q-card>
            <q-card-section>
              <div class="text-h1">Prometheus</div>
            </q-card-section>
            <q-card-section>
              <div class="text-h5">Votre rendez-vous</div>
            </q-card-section>
            <q-card-section>
              <!-- <div><strong>Invité :</strong> {{ appointment.email }}</div> -->
              <div><strong>Hôte :</strong> {{ appointment.host }}</div>
              <div><strong>Ecole :</strong> {{ appointment.school }}</div>
              <div>
                <strong>Date et Heure :</strong> {{ appointment.start_time.slice(11, 16) }}
              </div>
              <div>
                <strong>Heure de fin :</strong> {{ appointment.end_time.slice(11, 16) }}
              </div>
            </q-card-section>
          </q-card>
            <div class="q-mt-md row justify-around">
            <q-btn v-if="appointment.status === 'false'"
              label="Confirmer le rendez-vous"
              color="secondary"
              class="full-width q-mx-sm"
              @click="confirmingAppointment(token)"
            />
            <div class="q-mx-sm" style="height: 20px;"></div>
            <q-btn v-if="appointment.status === 'true'"
              label="Annuler le rendez-vous"
              color="negative"
              class="full-width q-mx-sm"
              @click="canceledAppointment(appointment.id)"
            />
            </div>
        </div>
      </q-page>
    </q-page-container>
  </q-layout>
</template>

<script>
import { useQuasar } from "quasar";
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { confirmAppointment, cancelAppointment } from "../services/appointment";

export default {
  name: "ValidateAppointment",
  props: {
    token: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const $q = useQuasar();
    const router = useRouter();
    const loading = ref(true);
    const appointment = ref(null);
    const server = import.meta.env.VITE_BASE_URL;

    async function fetchAppointments() {
      try {
        const response = await fetch(
          `${server}/validate-appointment?token=${props.token}`,
          {
            method: "GET",
          }
        );
        if (!response.ok) {
          throw new Error("Network response was not ok");
        }
        const data = await response.json();
        appointment.value = data;
      } catch (error) {
        console.error("Failed to fetch appointment details:", error);
        router.push("/error");
      } finally {
        loading.value = false;
      }
    }

    async function canceledAppointment () {
      try {
          const response = await fetch(
            `${server}/validate-appointment?token=${props.token}`,
            {
              method: "DELETE",
            }
          );
          if (response.status === 200) {
            router.push("/");
            $q.notify({
              type: "positive",
              message: "Rendez-vous supprimé avec succès !",
              position: "center",
            });
          }
        } catch (error) {
          console.error("Erreur lors de la suppression du rendez-vous:", error);
        }
    }

    const confirmingAppointment = async (token) => {
      try {
        const response = await confirmAppointment(token);
        if (response.status === 200) {
          router.push("/");
          $q.notify({
            type: "positive",
            message: "Rendez-vous confirmé avec succès !",
            position: "center",
          });
        }
      } catch (error) {
        console.error("Erreur lors de la confirmation du rendez-vous:", error);
      }
    };

    const goToHome = () => {
      router.push("/");
    };

    onMounted(
      fetchAppointments,
    );

    return {
      loading,
      appointment,
      goToHome,
      confirmingAppointment,
      canceledAppointment,
    };
  },
};
</script>

<style lang="scss" scoped>
.my-class {
  background-color: $primary;
  color: $policy;
}

.q-page {
  max-width: 800px;
  margin: 0 auto;
}
</style>
