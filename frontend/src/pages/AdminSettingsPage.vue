<template>
  <q-page class="q-pa-md">
    <q-card class="full-height">
      <q-card-section class="q-gutter-md row no-wrap" style="height: 85vh">
        <!-- Colonne gauche -->
        <div class="col-3 q-pa-sm">
          <!-- Ecoles -->
          <q-list
            bordered
            class="q-mb-md"
            style="max-height: 30%; overflow-y: auto"
          >
            <q-item>
              <q-item-label header>Ecoles</q-item-label>
              <q-item-section>
                <q-input
                  v-model="searchSchool"
                  dense
                  outlined
                  placeholder="Rechercher"
                  clearable
                  debounce="300"
                />
              </q-item-section>
              <q-item-section side>
                <q-btn
                  flat
                  round
                  dense
                  icon="add"
                  color="secondary"
                  @click="addSchool"
                />
              </q-item-section>
            </q-item>
            <q-skeleton v-if="!schools" type="rect" />
            <q-item
              v-else
              v-for="school in filteredSchoolsBySearch"
              :key="school.id"
              clickable
              @click="selectSchool(school)"
            >
              <q-item-section>{{ school.name }}</q-item-section>
            </q-item>
          </q-list>
          <!-- Agents -->
          <q-list
            bordered
            class="q-mb-md"
            style="max-height: 70%; overflow-y: auto"
          >
            <q-item>
              <q-item-label header>Agents</q-item-label>
              <q-item-section>
                <q-input
                  v-model="searchAgent"
                  dense
                  outlined
                  placeholder="Rechercher"
                  clearable
                  debounce="300"
                />
              </q-item-section>
              <q-item-section side>
                <q-btn
                  flat
                  round
                  dense
                  icon="add"
                  color="secondary"
                  @click="addAgent"
                />
              </q-item-section>
            </q-item>
            <q-skeleton v-if="!agents" type="rect" />
            <q-item
              v-else
              v-for="agent in filteredAgentsBySearch"
              :key="agent.id"
              clickable
              @click="selectAgent(agent)"
            >
              <q-item-section
                >{{ agent.given_name }} {{ agent.family_name }}</q-item-section
              >
            </q-item>
          </q-list>
        </div>
        <!-- Colonne droite -->
        <div class="col q-pa-sm">
          <!-- ECOLES -->
          <q-card flat bordered v-if="selectedSchool">
            <q-card-section class="row items-center justify-between">
              <div class="text-h6">{{ selectedSchool.name }}</div>
              <q-btn
                flat
                dense
                icon="delete"
                color="negative"
                @click="deleteSchool(selectedSchool)"
              />
            </q-card-section>

            <q-card-section>
              <q-input
                v-model="selectedSchool.id"
                label="ID"
                outlined
                class="q-mb-sm"
              />
              <q-input
                v-model="selectedSchool.name"
                label="Nom"
                outlined
                class="q-mb-sm"
              />
              <q-input
                v-model="selectedSchool.director_name"
                label="Directeur"
                outlined
                class="q-mb-sm"
              />
            </q-card-section>
          </q-card>

          <!-- Agent sélectionné -->
          <q-card v-if="selectedAgent" flat bordered class="q-mt-md">
            <q-card-section class="row items-center justify-between">
              <div class="text-subtitle1">
                {{ selectedAgent.given_name }} {{ selectedAgent.family_name }}
              </div>
              <q-btn
                flat
                dense
                icon="delete"
                color="negative"
                @click="deleteAgent(selectedAgent)"
              />
            </q-card-section>

            <q-card-section>
              <q-input
                v-model="selectedAgent.id"
                label="ID"
                outlined
                class="q-mb-sm"
              />
              <q-input
                v-model="selectedAgent.unique_name"
                label="Email"
                outlined
              />

              <!-- <div class="text-subtitle1 q-mt-md q-mb-sm">
                Ressources associées
              </div>
              <q-list bordered>
                <q-item>
                  <q-item-section>{{ test }}</q-item-section>
                </q-item>
              </q-list> -->
            </q-card-section>
          </q-card>
        </div>
      </q-card-section>
    </q-card>

    <!-- ECOLE -->
    <!-- q-dialog pour ajouter une école -->
    <q-dialog v-model="showAddSchoolDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Ajouter une école</div>
        </q-card-section>
        <q-card-section>
          <q-input v-model="newSchool.name" label="Nom" outlined dense />
          <q-select
            v-model="newSchool.director"
            use-input
            input-debounce="300"
            label="Rechercher une école"
            :options="emailSuggestions"
            @filter="(val, update, abort) => filterAzureEmail(val, update, abort)"
            option-label="mail"
            option-value="mail"
            emit-value
            map-options
            outlined
            dense
            class="q-mb-md"
            :popup-content-class="'q-select-popup'"
            @update:model-value="onDirectorSelect()"
          />
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showAddSchoolDialog = false"
          />
          <q-btn
            flat
            label="Ajouter"
            color="secondary"
            @click="confirmAddSchool"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    <!-- q-dialog pour supprimer une école -->
    <q-dialog v-model="showDeleteSchoolDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Supprimer l'école {{ selectedSchool.name }}</div>
        </q-card-section>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cette école ?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteSchoolDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteSchool"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    <!-- AGENT -->
    <!-- q-dialog pour ajouter un agent -->
    <q-dialog v-model="showAddAgentDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Ajouter un agent</div>
        </q-card-section>
        <q-card-section>
            <q-select
            v-model="newAgentSelection"
            use-input
            input-debounce="300"
            label="Rechercher un agent"
            :options="emailSuggestions"
            @filter="(val, update, abort) => filterAzureEmail(val, update, abort)"
            option-label="mail"
            option-value="mail"
            emit-value
            map-options
            outlined
            dense
            class="q-mb-md"
            :popup-content-class="'q-select-popup'"
            @update:model-value="onAgentSelect"
          />
            <q-input
            v-model="newAgent.unique_name"
            label="Email de l'agent (nom unique)"
            outlined
            dense
            readonly
            disable
            />
            <q-input
            v-model="newAgent.given_name"
            label="Prénom de l'agent"
            outlined
            dense
            class="q-mb-md"
            readonly
            disable
            />
            <q-input
            v-model="newAgent.family_name"
            label="Nom de l'agent"
            outlined
            dense
            class="q-mb-md"
            readonly
            disable
            />
            <q-input
            v-model="newAgent.schoolID"
            label="Ecole de l'agent"
            outlined
            dense
            class="q-mb-md"
            />
        </q-card-section>
        <q-card-actions align="right">
            <q-btn
            flat
            label="Annuler"
            color="negative"
            @click="showAddAgentDialog = false; emailSuggestions = []; newAgentSelection = null"
          />
          <q-btn
            flat
            label="Ajouter"
            color="secondary"
            @click="confirmAddAgent"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    <!-- q-dialog pour supprimer un agent -->
    <q-dialog v-model="showDeleteAgentDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Supprimer l'agent {{ selectedAgent.name }}</div>
        </q-card-section>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cet agent ?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteAgentDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteAgent"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script>
import { fetchSchools, insertSchool, deleteSchool } from "src/services/school";
import { fetchUsers, deleteUser, insertUser } from "src/services/user";
import { fetchUserSchoolResourceBySchoolID } from "src/services/userSchoolResource";

export default {
  data() {
    return {
      newSchool: {
        name: "",
        director: "",
        address: "",
        phone: "",
        email: "",
      },
      showSchoolDialog: false,
      showAddSchoolDialog: false,
      showDeleteSchoolDialog: false,

      newRessource: { name: "", description: "", duration: "" },
      showRessourceDialog: false,
      showAddRessourceDialog: false,
      showDeleteRessourceDialog: false,

      newAgent: { name: "", email: "", schoolID: "" },
      showAgentDialog: false,
      showAddAgentDialog: false,
      showDeleteAgentDialog: false,

      searchAgent: "",
      searchSchool: "",

      splitterModel: 300,

      selectedResource: null,
      selectedAgent: null,
      selectedSchool: null,

      schools: null,
      resourceTypes: null,
      agents: null,
      emailSuggestions: [],
    };
  },

  mounted() {
    this.getSchools();
    this.getAgents();
  },

  computed: {
    // methode pour filtrer les agents en fonction de la ressource sélectionnée
    filteredAgents() {
      if (!this.selectedResource) return [];
      return this.agents.filter((agent) =>
        agent.resourceIds.includes(this.selectedResource.id)
      );
    },
    // methode pour filtrer les ressources en fonction de l'agent sélectionné
    resourcesForAgent() {
      if (!this.selectedAgent) return [];
      return this.resourceTypes.filter((res) =>
        this.selectedAgent.resourceIds.includes(res.id)
      );
    },
    // methode pour filtrer les agents en fonction de la recherche (barre de recherche)
    filteredAgentsBySearch() {
      if (!this.searchAgent) return this.agents;
      const search = this.searchAgent.toLowerCase();
      return this.agents.filter(
        (agent) =>
          (agent.given_name &&
            agent.given_name.toLowerCase().includes(search)) ||
          (agent.unique_name &&
            agent.unique_name.toLowerCase().includes(search)) ||
          (agent.family_name &&
            agent.family_name.toLowerCase().includes(search))
      );
    },
    // methode pour filtrer les écoles en fonction de la recherche (barre de recherche)
    filteredSchoolsBySearch() {
      if (!this.searchSchool) return this.schools;
      const search = this.searchSchool.toLowerCase();
      return this.schools.filter((school) =>
        school.name.toLowerCase().includes(search)
      );
    },
  },
  methods: {
    // Méthodes pour filtrer les agents en fonction de la recherche
    onAgentSelect(val) {
      console.log("Valeur sélectionnée:", val);
      const selected =
      this.emailSuggestions &&
      this.emailSuggestions.find(
        (item) => item.mail === val || item.mail === this.newAgentSelection
      );
      if (selected) {
      this.newAgent.unique_name = selected.mail;
      this.newAgent.given_name = selected.given_name || "";
      this.newAgent.family_name = selected.family_name || "";
      } else if (typeof val === "string") {
      this.newAgent.unique_name = val;
      }
    },
    onDirectorSelect(val) {
      console.log("Valeur sélectionnée:", val);
      newSchool.director = val;
      // const selected =
      // this.emailSuggestions &&
      // this.emailSuggestions.find(
      //   (item) => item.mail === val || item.mail === this.newAgentSelection
      // );
      // if (selected) {
      // this.newAgent.unique_name = selected.mail;
      // this.newAgent.given_name = selected.given_name || "";
      // this.newAgent.family_name = selected.family_name || "";
      // } else if (typeof val === "string") {
      // this.newAgent.unique_name = val;
      // }
    },
    // Méthode pour filtrer les emails Azure AD
    async filterAzureEmail(val, update) {
      if (!val) {
        update(() => {
          this.emailSuggestions = [];
        });
        return;
      }
      try {
        const response = await fetch(`${import.meta.env.VITE_BASE_URL}get-azure-users?search=${encodeURIComponent(val)}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        });
        const data = await response.json();
        this.emailSuggestions = data;
        update();
      } catch (error) {
        console.error("Erreur Azure AD:", error);
        update(() => {
          this.emailSuggestions = [];
        });
      }
    },
    selectSchool(school) {
      this.selectedSchool = school;
      this.selectedResource = null;
      this.selectedAgent = null;
    },
    selectResource2(resource) {
      this.selectedResource = resource;
      this.selectedAgent = null;
    },
    selectResource(resource) {
      this.selectedSchool = null;
      this.selectedResource = resource;
      this.selectedAgent = null;
    },
    selectAgent(agent) {
      this.selectedSchool = null;
      this.selectedResource = null;
      this.selectedAgent = agent;
    },

    // CRUD école
    // TODO : Changer les methodes actuelles par des appels d'API
    // Recuperation de la liste des écoles
    getSchools() {
      fetchSchools()
        .then((response) => {
          this.schools = response;
        })
        .catch((error) => {
          console.error("Erreur lors de la récupération des écoles", error);
        });
    },
    // Ajout d'une école
    addSchool() {
      this.newSchool = {
        name: "",
        director: "",
        address: "",
        phone: "",
        email: "",
      };
      this.showAddSchoolDialog = true;
    },
    confirmAddSchool() {
      if (
        !this.newSchool.name ||
        !this.newSchool.director
      ) {
        this.$q.notify({
          message: "Veuillez remplir tous les champs.",
          color: "red",
        });
        return;
      }
      insertSchool(this.newSchool)
        .then((response) => {
          this.showAddSchoolDialog = false;
          this.$q.notify({
            message: "Ecole ajouté avec succès.",
            color: "green",
            position: "center",
          });
          this.reloadSchools();
        })
        .catch((error) => {
          console.error("Erreur lors de l'ajout de l'école", error);
          this.$q.notify({
            message: "Erreur lors de l'ajout de l'école.",
            color: "red",
            position: "center",
          });
        });
    },
    // Suppression d'une école
    deleteSchool(school) {
      this.showDeleteSchoolDialog = true;
      this.selectedSchool = school;
    },
    confirmDeleteSchool() {
      if (!this.selectedSchool) return;
      const school = this.selectedSchool;
      deleteSchool(school.id, true)
        .then(() => {
          this.$q.notify({
            message: `L'école ${school.name}a été supprimée.`,
            color: "green",
            position: "center",
          });
          this.reloadSchools();
        })
        .catch((error) => {
          console.error("Error deleting school:", error);
          this.$q.notify({
            message: "Erreur lors de la suppression de l'école.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.selecedtSchool = null;
          this.showDeleteSchoolDialog = false;
        });
    },
    // Rechargement de la liste des écoles
    reloadSchools() {
      fetchSchools()
        .then((response) => {
          this.schools = response;
        })
        .catch((error) => {
          console.error("Erreur lors de la récupération des écoles", error);
        });
    },

    // CRUD ressource
    // TODO : Changer les methodes actuelles par des appels d'API
    // Recuperation de la liste des ressources
    // getResources() {},
    // // Ajout d'une ressource
    // addResource() {
    //   const name = prompt("Nom de la ressource :");
    //   if (!name) return;
    //   const description = prompt("Description de la ressource :");
    //   const duration = prompt("Durée d'un rendez-vous :");
    //   const newRes = {
    //     id: this.resourceTypes.length + 1,
    //     name,
    //     description,
    //     duration,
    //   };
    //   this.resourceTypes.push(newRes);
    // },
    // confirmAddRessource() {
    //   if (!this.newRessource.name || !this.newRessource.description) {
    //     this.$q.notify({
    //       message: "Veuillez remplir tous les champs.",
    //       color: "red",
    //     });
    //     return;
    //   }
    //   const newResource = {
    //     id: this.resourceTypes.length + 1,
    //     name: this.newRessource.name,
    //     description: this.newRessource.description,
    //     duration: this.newRessource.duration,
    //   };
    //   this.resourceTypes.push(newResource);
    //   this.showAddRessourceDialog = false;
    //   this.$q.notify({
    //     message: "Ressource ajoutée avec succès.",
    //     color: "green",
    //   });
    // },
    // // Suppression d'une ressource
    // deleteResource(resource) {
    //   this.showDeleteRessourceDialog = true;
    //   this.selectedResource = resource;
    // },
    // confirmDeleteRessource() {
    //   if (!this.selectedResource) return;
    //   const resource = this.selectedResource;
    //   this.resourceTypes = this.resourceTypes.filter(
    //     (r) => r.id !== resource.id
    //   );
    //   this.selectedResource = null;
    //   this.$q.notify({
    //     message: `La ressource ${resource.name} a été supprimée.`,
    //     color: "green",
    //   });
    //   this.showDeleteRessourceDialog = false;
    // },

    // CRUD agent
    // TODO : Changer les methodes actuelles par des appels d'API
    // Recuperation de la liste des agents
    getAgents() {
      fetchUsers()
        .then((response) => {
          this.agents = response;
        })
        .catch((error) => {
          console.error("Erreur lors de la récupération des agents", error);
        });
    },
    // Ajout d'un agent
    addAgent() {
      this.newAgent = { name: "", email: "" };
      this.showAddAgentDialog = true;
    },
    confirmAddAgent() {
      const agent = this.newAgent;
      if (
        !this.newAgent.given_name ||
        !this.newAgent.family_name ||
        !this.newAgent.unique_name ||
        !this.newAgent.schoolID
      ) {
        this.$q.notify({
          message: "Veuillez remplir tous les champs.",
          color: "red",
        });
        return;
      }
      insertUser(agent, agent.schoolID)
        .then(() => {
          this.$q.notify({
            message: `L'agent ${agent.given_name} ${agent.family_name} a été ajouté avec succès.`,
            color: "green",
            position: "center",
          });
        })
        .catch((error) => {
          if (error.response && error.response.data) {
            console.error("Error creating agent:", error.response.data);
          } else {
            console.error("Error creating agent:", error.message || error);
          }
          this.$q.notify({
            message: "Erreur lors de la création de l'agent.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.newAgent = { given_name: "", family_name: "", unique_name: "" };
          this.showAddAgentDialog = false;
          this.reloadAgents();
        });
    },
    // Suppression d'un agent
    deleteAgent(agent) {
      this.showDeleteAgentDialog = true;
      this.selectedAgent = agent;
    },
    confirmDeleteAgent() {
      if (!this.selectedAgent) return;
      const agent = this.selectedAgent;
      const source = "admin";
      deleteUser(agent.id, source)
      .then(() => {
        this.$q.notify({
        message: `L'agent ${agent.given_name} ${agent.family_name} a été supprimé.`,
        color: "green",
        position: "center",
        });
        this.getAgents(); // Met à jour la liste des agents après suppression
      })
      .catch((error) => {
        console.error("Error deleting agent:", error);
        this.$q.notify({
        message: "Erreur lors de la suppression de l'agent.",
        color: "red",
        position: "center",
        });
      })
      .finally(() => {
        this.selectedAgent = null;
        this.showDeleteAgentDialog = false;
      });
    },
    reloadAgents() {
      if (!this.selectedSchool) return;
      fetchUserSchoolResourceBySchoolID(this.selectedSchool.id)
      .then(({ agents }) => {
        this.agents = agents;
      })
      .catch((error) => {
        console.error("Error reloading agents:", error);
        this.$q.notify({
        message: "Erreur lors du rechargement des agents.",
        color: "red",
        position: "center",
        });
      });
    },

    // CRUD liaison école-ressource
    // TODO : Changer les methodes actuelles par des appels d'API
    // Ajout d'une ressource à une école
    addRessourceToSchool(resource) {
      if (!this.selectedSchool) return;
      if (!this.selectedSchool.resourceIds) {
        this.selectedSchool.resourceIds = [];
      }
      if (!this.selectedSchool.resourceIds.includes(resource.id)) {
        this.selectedSchool.resourceIds.push(resource.id);
        this.$q.notify({
          message: `La ressource ${resource.name} a été ajoutée à l'école ${this.selectedSchool.name}`,
          color: "green",
        });
      }
      this.showRessourceDialog = false;
    },
    // Suppression d'une ressource d'une école
    removeResourceFromSchool(resource) {
      if (!this.selectedSchool) return;
      const idx = this.selectedSchool.resourceIds.indexOf(resource.id);
      if (idx !== -1) {
        this.selectedSchool.resourceIds.splice(idx, 1);
        this.$q.notify({
          message: `La ressource ${resource.name} a été supprimée de l'école ${this.selectedSchool.name}`,
          color: "green",
        });
      }
    },
    // affichage du pop-up pour ajouter une ressource à une école
    openRessourceSelectionDialog() {
      if (!this.selectedSchool) return;
      this.showRessourceDialog = true;
    },

    // CRUD liaison agent-ressource
    // TODO : Changer les methodes actuelles par des appels d'API
    // Ajout d'un agent à une ressource
    addAgentToResource(agent) {
      if (!this.selectedResource) return;
      if (!agent.resourceIds.includes(this.selectedResource.id)) {
        agent.resourceIds.push(this.selectedResource.id);
        this.$q.notify({
          message: `L'agent ${agent.name} a été ajouté à la ressource ${this.selectedResource.name}`,
          color: "green",
        });
      }
      this.showAgentDialog = false;
    },
    // Suppression d'un agent d'une ressource
    removeAgentFromResource(agent, resource) {
      const idx = agent.resourceIds.indexOf(resource.id);
      if (idx !== -1) {
        agent.resourceIds.splice(idx, 1);
        this.$q.notify({
          message: `L'agent ${agent.name} a été supprimé de la ressource ${resource.name}`,
          color: "green",
        });
      }
    },
    // affichage du pop-up pour ajouter un agent à une ressource
    openAgentSelectionDialog() {
      if (!this.selectedResource) return;
      this.showAgentDialog = true;
    },
  },
};
</script>
