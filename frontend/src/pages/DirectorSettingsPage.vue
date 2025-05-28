<template>
  <q-page class="q-pa-md">
    <q-card class="full-height">
      <q-card-section class="q-gutter-md row no-wrap" style="height: 85vh">
        <!-- Colonne gauche -->
        <div class="col-3 q-pa-sm">
          <!-- Ecoles -->
          <q-select
            v-if="schools.length > 1"
            v-model="selectedSchool"
            :options="schools"
            option-value="id"
            option-label="name"
            label="Changer d'école"
            outlined
            dense
            class="q-mb-md"
            @update:model-value="changeSchool(selectedSchool.id)"
          />
          <!-- Resources -->
          <q-list
            bordered
            class="q-mb-md"
            style="max-height: 40%; overflow-y: auto"
          >
            <q-item>
              <q-item-label header>Resources</q-item-label>
              <q-item-section>
                <q-input
                  v-model="searchResource"
                  dense
                  outlined
                  placeholder="Rechercher"
                  clearable
                  debounce="300"
                  @input="filterResources"
                />
              </q-item-section>
              <q-item-section side>
                <q-btn
                  flat
                  round
                  dense
                  icon="add"
                  color="secondary"
                  @click="addResource"
                />
              </q-item-section>
            </q-item>
            <q-item
              v-for="resource in (filteredResourcesBySearch || []).filter(r => r.visible).sort(
              (a, b) => a.id - b.id
              )"
              :key="resource.id"
              clickable
              @click="selectResource(resource)"
            >
              <q-item-section>{{ resource.description }}</q-item-section>
            </q-item>
          </q-list>
          <!-- Agents -->
          <q-list
            bordered
            class="q-mb-md"
            style="max-height: 50%; overflow-y: auto"
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
                  @input="filteredAgents"
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
            <q-item
              v-for="agent in (filteredAgentsBySearch || []).sort(
                (a, b) => a.id - b.id
              )"
              :key="agent.id || agent.unique_name"
              clickable
              @click="selectAgent(agent)"
            >
              <q-item-section>{{
                agent.given_name + " " + agent.family_name
              }}</q-item-section>
            </q-item>
          </q-list>
        </div>

        <!-- Colonne droite -->
        <div class="col q-pa-sm">
          <!-- Resource selectionnée -->
          <q-card flat bordered v-if="selectedResource">
            <q-card-section class="row items-center justify-between">
              <q-banner header>{{ selectedResource.description }}</q-banner>
              <q-btn
                flat
                dense
                icon="delete"
                color="negative"
                @click="deleteResource(selectedResource)"
              />
            </q-card-section>
            <q-card-section>
              <q-input v-model="selectedResource.name" label="Nom" outlined />
              <q-input
                v-model="selectedResource.description"
                label="Description"
                outlined
              />
              <q-input
                v-model="selectedResource.duration"
                label="Durée"
                outlined
                class="q-mb-md"
              />
              <q-btn
                class="q-mb-sm full-width"
                label="Modifier"
                color="secondary"
                @click="updateResource()"
              />
            </q-card-section>
            <!-- Agents associés à la ressource sélectionnée -->
            <q-card-section v-if="selectedResource">
              <q-banner header>Agents associés</q-banner>
              <q-list bordered>
                <q-item
                  v-for="agent in filteredAgents"
                  :key="agent.id"
                  clickable
                >
                  <q-item-section>{{
                    agent.given_name + " " + agent.family_name
                  }}</q-item-section>
                  <q-item-section side>
                    <q-btn
                      flat
                      round
                      dense
                      icon="delete"
                      color="negative"
                      @click.stop="deleteAgentFromResource(agent)"
                    />
                  </q-item-section>
                </q-item>
              </q-list>
              <q-btn
                class="q-mt-md full-width"
                label="Ajouter un agent"
                color="secondary"
                @click="openAgentSelectionDialog"
              />
            </q-card-section>
          </q-card>

          <!-- Agent sélectionné -->
          <q-card flat bordered v-if="selectedAgent && !selectedResource">
            <q-card-section class="row items-center justify-between">
              <div class="text-h6">
                {{ selectedAgent?.given_name || "Prénom inconnu" }}
                {{ selectedAgent?.family_name || "Nom inconnu" }}
              </div>
            </q-card-section>

            <q-card-section>
              <q-input
                v-model="selectedAgent.family_name"
                label="Nom"
                outlined
              />
              <q-input
                v-model="selectedAgent.given_name"
                label="Prénom"
                outlined
              />
              <q-input
                v-model="selectedAgent.unique_name"
                label="Email"
                outlined
              />
            </q-card-section>
            <q-card-section>
              <q-banner class="q-mb-sm">Resource(s) associés(s)</q-banner>
              <q-list bordered>
                <q-item
                  v-for="resource in filteredResources"
                  :key="resource.id"
                  clickable
                >
                  <q-item-section>{{ resource.description }}</q-item-section>
                  <q-item-section side>
                    <q-btn
                      flat
                      round
                      dense
                      icon="delete"
                      color="negative"
                      @click.stop="
                        deleteAgentFromResource2(resource, this.selectedAgent)
                      "
                    />
                  </q-item-section>
                </q-item>
              </q-list>
            </q-card-section>
          </q-card>
        </div>
      </q-card-section>
    </q-card>

    <!-- Q-DIALOG -->
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

    <!-- RESSOURCE -->
    <!-- Q-DIALOG pour ajouter une ressource -->
    <q-dialog v-model="showAddRessourceDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Ajouter une ressource</div>
        </q-card-section>
        <q-card-section>
          <q-input
            v-model="newRessource.name"
            label="Nom de la ressource"
            outlined
            dense
            class="q-mb-md"
          />
          <q-input
            v-model="newRessource.description"
            label="Description de la ressource"
            outlined
            dense
            class="q-mb-md"
          />
          <q-input
            v-model="newRessource.duration"
            label="Durée de la ressource"
            outlined
            dense
            class="q-mb-md"
          />
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showAddRessourceDialog = false"
          />
          <q-btn
            flat
            label="Ajouter"
            color="secondary"
            @click="confirmAddRessource"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    <!-- Q-DIALOG pour supprimer une ressource -->
    <q-dialog v-model="showDeleteRessourceDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Supprimer la ressource</div>
        </q-card-section>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cette ressource ?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteRessourceDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteRessource"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>


    <q-dialog v-model="showDeleteRessourceDialog">
      <q-card>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cette ressource de l'école ?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteRessourceDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteRessource"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- RESOURCE-AGENT -->
    <!-- Q-DIALOG pour ajouter un agent à une ressource -->
    <q-dialog v-model="showAddAgentToResourceDialog">
      <q-card>
        <q-card-section>
          <div class="text-h5">Sélectionnez un agent</div>
        </q-card-section>
        <q-list>
          <q-item
            v-for="agent in agents"
            :key="agent.id"
            clickable
            @click="addAgentToResource(agent)"
          >
            <q-item-section>{{
              agent.given_name + " " + agent.family_name
            }}</q-item-section>
          </q-item>
        </q-list>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showAddAgentToResourceDialog = false"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    <!-- Q-DIALOG pour supprimer un agent d'une ressource -->
    <q-dialog v-model="showDeleteAgentFromResourceDialog">
      <q-card>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cet agent de la resource?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteAgentFromResourceDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteAgentFromResource"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <q-dialog v-model="showDeleteAgentFromResourceDialog2">
      <q-card>
        <q-card-section>
          Êtes-vous sûr de vouloir supprimer cet agent de la resource?
        </q-card-section>
        <q-card-actions align="right">
          <q-btn
            flat
            label="Annuler"
            color="primary"
            @click="showDeleteAgentFromResourceDialog = false"
          />
          <q-btn
            flat
            label="Supprimer"
            color="negative"
            @click="confirmDeleteAgentFromResource2"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script>
import { getUser } from "src/services/api";
import { fetchSchools } from "src/services/school";
import {
  createResource,
  fetchResources,
  updateResource,
  deleteResource,
} from "src/services/resource";
import {
  fetchUserSchoolResourceBySchoolID,
  deleteUserSchoolResource,
  updateUserSchoolResource,
} from "src/services/userSchoolResource";
import { insertUser } from "src/services/user";

export default {
  data() {
    return {
      newRessource: { name: "", description: "", duration: "" },
      showRessourceDialog: false,
      showAddRessourceDialog: false,
      showDeleteRessourceDialog: false,

      newAgent: { given_name: "", family_name: "", unique_name: "" },
      showAddAgentToResourceDialog: false,
      showAddAgentDialog: false,
      showDeleteAgentDialog: false,
      showDeleteAgentFromResourceDialog: false,
      showDeleteAgentFromResourceDialog2: false,

      searchAgent: "",
      searchResource: "",

      splitterModel: 300,

      selectedSchool: null,
      selectedResource: null,
      selectedAgent: null,
      selResource: null,

      schools: [],
      resources: [],
      agents: [],
      emailSuggestions: [],
      newAgentSelection: null,
    };
  },
  computed: {
    // methode pour filtrer les agents en fonction de la ressource sélectionnée
    filteredAgents() {
      if (!this.selectedResource) return [];
      const filtered = this.agents.filter((agent) => {
        if (Array.isArray(agent.resourceId)) {
          return agent.resourceId.includes(this.selectedResource.id);
        }
        return agent.resourceId === this.selectedResource.id;
      });
      return filtered;
    },
    // Filtrer les ressources associées à l'agent sélectionné
    filteredResources() {
      if (!this.selectedAgent) return [];
      const agent = this.agents.find((a) => a.id === this.selectedAgent.id);
      if (!agent || !agent.resourceId) return [];

      // Si resourceId est un tableau, filtrez les ressources correspondantes
      const resourceIds = Array.isArray(agent.resourceId)
        ? agent.resourceId
        : [agent.resourceId];

      return this.resources.filter((resource) =>
        resourceIds.includes(resource.id)
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
    // methode pour filtrer les ressources en fonction de la ressource sélectionnée
    filteredResource() {
      if (!this.selectedResource) return [];
      return this.agents.filter((agent) =>
        agent.resource_id.includes(this.selectedResource.id)
      );
    },
    // methode pour filtrer les ressources en fonction de la recherche (barre de recherche)
    filteredResourcesBySearch() {
      if (!this.searchResource) return this.resources;
      const search = this.searchResource.toLowerCase();
      return this.resources.filter((resource) =>
        resource.resource_description.toLowerCase().includes(search)
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
    filterResources() {
      const search = this.searchResource.toLowerCase();
      this.filteredResourcesBySearch = this.resources.filter((resource) =>
        resource.name.toLowerCase().includes(search)
      );
    },
    selectResource2(resource) {
      this.selectedResource = resource;
      this.selectedAgent = null;
    },
    selectResource(resource) {
      this.selectedResource = resource;
      this.selectedAgent = null;
    },
    selectAgent(agent) {
      this.selectedAgent = agent;
      this.selectedResource = null;
    },
    // Changer d'école
    changeSchool(schoolID) {
      if (!schoolID) return;
      this.selectedSchool = this.schools.find(
        (school) => school.id === schoolID
      );
      this.selectedResource = null;
      this.selectedAgent = null;
      this.reloadResources();
      this.reloadAgents();
    },
    // CRUD ressource
    // Ajout d'une ressource
    addResource() {
      this.newRessource = {
        name: "",
        description: "",
        school: this.selectedSchool.id,
        duration: "",
      };
      this.showAddRessourceDialog = true;
    },
    // Confirmation de l'ajout d'une ressource
    confirmAddRessource() {
      if (
      !this.newRessource.name ||
      !this.newRessource.description ||
      !this.newRessource.duration
      ) {
      this.$q.notify({
        message: "Veuillez remplir tous les champs.",
        color: "red",
      });
      return;
      }
      const payload = {
      ...this.newRessource,
      duration: parseInt(this.newRessource.duration, 10)
      };
      payload.visible = true;
      createResource(payload)
      .then(() => {
        this.showAddRessourceDialog = false;
        this.$q.notify({
        message: "Ressource ajoutée avec succès.",
        color: "green",
        position: "center",
        });
        this.reloadResources();
      })
      .catch((error) => {
        console.error("Error creating resource:", error);
        this.$q.notify({
        message: "Erreur lors de la création de la ressource.",
        color: "red",
        });
      });
    },
    // Suppression d'une ressource
    deleteResource(resource) {
      this.showDeleteRessourceDialog = true;
      this.selectedResource = resource;
    },
    confirmDeleteRessource() {
      if (!this.selectedResource) {
        this.$q.notify({
          message: "Aucune ressource sélectionnée.",
          color: "red",
          position: "center",
        });
        return;
      }
      // const resourceID = this.selectedResource.id;
      deleteResource(this.selectedResource.id)
        .then(() => {
          this.$q.notify({
            message: `La ressource ${this.selectedResource.name} a été supprimée.`,
            color: "green",
            position: "center",
          });
        })
        .catch((error) => {
          console.error("Error deleting resource:", error);
          this.$q.notify({
            message: "Erreur lors de la suppression de la ressource.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.selectedResource = null;
          this.showDeleteRessourceDialog = false;
          this.reloadResources();
        });
    },
    // Modification d'une ressource
    updateResource() {
      if (!this.selectedResource) {
        this.$q.notify({
          message: "Aucune ressource sélectionnée.",
          color: "red",
          position: "center",
        });
        return;
      }
      const resource = this.selectedResource;
      if (!resource.name || !resource.description || !resource.duration) {
        this.$q.notify({
          message: "Veuillez remplir tous les champs.",
          color: "red",
        });
        return;
      }
      resource.duration = parseInt(resource.duration, 10);
      updateResource(resource)
        .then(() => {
          this.$q.notify({
            message: `La ressource ${resource.name} a été modifiée.`,
            color: "green",
            position: "center",
          });
        })
        .catch((error) => {
          console.error("Error updating resource:", error);
          this.$q.notify({
            message: "Erreur lors de la modification de la ressource.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.selectedResource = null;
          this.reloadResources();
        });
    },
    reloadResources() {
      if (!this.selectedSchool) return;
      fetchResources(this.selectedSchool.id)
        .then((resources) => {
          this.resources = resources;
        })
        .catch((error) => {
          console.error("Error reloading resources:", error);
          this.$q.notify({
            message: "Erreur lors du rechargement des ressources.",
            color: "red",
            position: "center",
          });
        });
    },

    // CRUD agent
    addAgent() {
      this.newAgent = { given_name: "", family_name: "", unique_name: "" };
      this.showAddAgentDialog = true;
    },
    confirmAddAgent() {
      if (
        !this.newAgent.given_name ||
        !this.newAgent.family_name ||
        !this.newAgent.unique_name
      ) {
        this.$q.notify({
          message: "Veuillez remplir tous les champs.",
          color: "red",
        });
        return;
      }
      const agent = this.newAgent;
      insertUser(agent, this.selectedSchool.id)
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
    // affichage du pop-up pour ajouter une ressource à une école
    openRessourceSelectionDialog() {
      if (!this.selectedSchool) return;
      this.showRessourceDialog = true;
    },

    // CRUD liaison agent-ressource
    // Ajout d'un agent à une ressource
    addAgentToResource(agent) {
      updateUserSchoolResource(
        agent.id,
        this.selectedSchool.id,
        this.selectedResource.id
      )
        .then(() => {
          this.$q.notify({
            message: `L'agent ${agent.given_name} ${agent.family_name} a été ajouté à la ressource ${this.selectedResource.description}.`,
            color: "green",
            position: "center",
          });
          this.reloadAgents();
          this.showAddAgentToResourceDialog = false;
        })
        .catch((error) => {
          console.error("Error adding agent to resource:", error);
          this.$q.notify({
            message: "Erreur lors de l'ajout de l'agent à la ressource.",
            color: "red",
          });
        });
    },
    // Suppression d'un agent d'une ressource (pour la liste des ressources)
    deleteAgentFromResource(agent) {
      this.showDeleteAgentFromResourceDialog = true;
      this.selectedAgent = agent;
    },
    confirmDeleteAgentFromResource() {
      if (!this.selectedResource) return;
      if (!this.selectedAgent) return;

      const agent = this.selectedAgent;

      deleteUserSchoolResource(
        agent.id,
        this.selectedSchool.id,
        this.selectedResource.id
      )
        .then(() => {
          this.$q.notify({
            message: `L'agent ${agent.given_name} ${agent.family_name} a été supprimé de la ressource ${this.selectedResource.description}.`,
            color: "green",
            position: "center",
          });
          this.reloadAgents();
          this.showDeleteAgentFromResourceDialog = false;
        })
        .catch((error) => {
          console.error("Error deleting agent from resource:", error);
          this.$q.notify({
            message:
              "Erreur lors de la suppression de l'agent de la ressource.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.selectedAgent = null;
        });
    },
    // Suppression d'un agent d'une ressource (pour la liste des agents)
    deleteAgentFromResource2(resource, agent) {
      this.showDeleteAgentFromResourceDialog2 = true;
      this.selectedAgent = agent;
      this.selResource = resource;
    },

    confirmDeleteAgentFromResource2() {
      if (!this.selResource) return;
      if (!this.selectedAgent) return;

      const agent = this.selectedAgent;

      deleteUserSchoolResource(
        agent.id,
        this.selectedSchool.id,
        this.selResource.id
      )
        .then(() => {
          this.$q.notify({
            message: `L'agent ${agent.given_name} ${agent.family_name} a été supprimé de la ressource ${this.selResource.description}.`,
            color: "green",
            position: "center",
          });
          this.reloadAgents();
          this.showDeleteAgentFromResourceDialog2 = false;
        })
        .catch((error) => {
          console.error("Error deleting agent from resource:", error);
          this.$q.notify({
            message:
              "Erreur lors de la suppression de l'agent de la ressource.",
            color: "red",
            position: "center",
          });
        })
        .finally(() => {
          this.selectedAgent = null;
          this.selResource = null;
        });
    },

    // DIALOG
    openAgentSelectionDialog() {
      if (!this.selectedResource) return;
      this.showAddAgentToResourceDialog = true;
    },
  },
  watch: {
  newAgentSelection(selected) {
    if (selected && typeof selected === "object") {
      this.newAgent.unique_name = selected.mail;
      this.newAgent.given_name = selected.given_name || "";
      this.newAgent.family_name = selected.family_name || "";
    } else {
      this.newAgent.unique_name = selected || "";
    }
  }
},

  mounted: async function () {
    try {
      this.schools = await fetchSchools(getUser().unique_name);
      this.selectedSchool = this.schools[0];
      this.resources = await fetchResources(this.selectedSchool.id);
      const { resources, agents } = await fetchUserSchoolResourceBySchoolID(
        this.selectedSchool.id
      );
      this.agents = agents;
    } catch (error) {
      console.error("Error fetching resources and agents:", error);
      this.$q.notify({
        type: "negative",
        message: "Erreur lors de la récupération des données.",
      });
    }
  },

};
</script>


