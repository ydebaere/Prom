const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function fetchResources(schoolID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + `resources?schoolID=${schoolID}`, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching types:", error);
    throw error;
  }
}

export async function createResource(resource) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + "resources", {
      method: "POST",
      headers,
      body: JSON.stringify(resource),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    return response;
  } catch (error) {
    console.error("Error adding type:", error);
    throw error;
  }
}

export async function updateResource(resource) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + `resources`, {
      method: "PUT",
      headers,
      body: JSON.stringify(resource),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching types:", error);
    throw error;
  }
}

export async function deleteResource(resource_id) {
  const url = `${server}resources?resourceID=${resource_id}`;

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
    return response;
  } catch (error) {
    console.error("Error deleting type:", error);
    throw error;
  }
}
