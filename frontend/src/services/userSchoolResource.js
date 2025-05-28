const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function fetchUserSchoolResource() {
  const url = `${server}user-school-resource?`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return {
      agents: data.agents,
      schools: data.schools,
      resources: data.resources,
    };
  } catch (error) {
    console.error("Error fetching userRole:", error);
    throw error;
  }
}

export async function fetchUserSchoolResourceByUserID(userID) {
  const url = `${server}user-school-resource?userID=${userID}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return {};
  } catch (error) {
    console.error("Error fetching userRole:", error);
    throw error;
  }
}

export async function fetchUserSchoolResourceBySchoolID(schoolID) {
  const url = `${server}user-school-resource?schoolID=${schoolID}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error(
        `Network response was not ok. Status: ${response.status}`
      );
    }

    const data = await response.json();

    if (!data || data.length === 0) {
      return {
        agents: [],
        resources: [],
      };
    }

    // Extraire les ressources avec leurs utilisateurs associés
    const resources = data.map((resource) => ({
      id: resource.resource_id,
      name: resource.resource_name,
      description: resource.resource_description,
      duration: resource.resource_duration,
      users: resource.users.map((user) => ({
        id: user.user_id,
        given_name: user.given_name,
        family_name: user.family_name,
        unique_name: user.unique_name,
      })),
    }));

    // Extraire les agents à partir des utilisateurs associés aux ressources
    const agentsMap = new Map(); // Utiliser une Map pour éviter les doublons d'agents
    data.forEach((resource) => {
      resource.users.forEach((user) => {
        if (!agentsMap.has(user.user_id)) {
          agentsMap.set(user.user_id, {
            id: user.user_id,
            given_name: user.given_name,
            family_name: user.family_name,
            unique_name: user.unique_name,
            resourceId: resource.resource_id,
            resource_name: resource.resource_name,
          });
        } else {
          // Ajouter la ressource à laquelle l'agent est lié
          const existingAgent = agentsMap.get(user.user_id);
          existingAgent.resourceId = [
            ...(Array.isArray(existingAgent.resourceId)
              ? existingAgent.resourceId
              : [existingAgent.resourceId]),
            resource.resource_id,
          ];
          existingAgent.resource_name = [
            ...(Array.isArray(existingAgent.resource_name)
              ? existingAgent.resource_name
              : [existingAgent.resource_name]),
            resource.resource_name,
          ];
        }
      });
    });

    const agents = Array.from(agentsMap.values());

    return {
      agents,
      resources,
    };
  } catch (error) {
    console.error("Error fetching userRole:", error);
    throw error;
  }
}

export async function fetchUserSchoolResourceByResourceID(resourceID) {
  const url = `${server}user-school-resource?resourceID=${resourceID}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return {
      agents: data.agents,
      schools: data.schools,
      resources: data.resources,
    };
  } catch (error) {
    console.error("Error fetching userRole:", error);
    throw error;
  }
}

export async function createUserSchoolResource(userID, schoolID, resourceID) {
  const url = `${server}users`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "POST",
      headers,
      body: JSON.stringify({
        userID,
        schoolID,
        resourceID,
      }),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error creating userRole:", error);
    throw error;
  }
}

export async function updateUserSchoolResource(userID, schoolID, resourceID) {
  const url = `${server}user-school-resource?userID=${userID}&schoolID=${schoolID}&resourceID=${resourceID}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "PUT",
      headers,
      body: JSON.stringify({
        userID,
        schoolID,
        resourceID,
      }),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    return response;
  } catch (error) {
    console.error("Error creating userRole:", error);
    throw error;
  }
}

export async function deleteUserSchoolResource(userID, schoolID, resourceID) {
  const url = `${server}user-school-resource?userID=${userID}&schoolID=${schoolID}&resourceID=${resourceID}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "DELETE",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error deleting userSchoolRole:", error);
    throw error;
  }
}
