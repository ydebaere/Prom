const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function fetchUsers() {
  try {
    if (!token) {
      throw new Error("No token found");
    }

    const response = await fetch(server + "users", {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching users:", error);
    throw error;
  }
}

export async function fetchUserByID(user_id) {
  const url = `${server}users?userID=${user_id}`;
  try {
    if (!token) {
      throw new Error();
    }
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return JSON.stringify(data);
  } catch (error) {
    console.error("Error fetching users:", error);
    throw error;
  }
}

export async function fetchAgents(schoolID, resourceID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const url = `${server}user-school-resource?schoolID=${schoolID}&resourceID=${resourceID}`;
    const response = await fetch(url, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching users:", error);
    throw error;
  }
}

export async function insertUser(user, schoolID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + `users?schoolID=${schoolID}`, {
      method: "POST",
      headers,
      body: JSON.stringify(user),
    });
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  } catch (error) {
    console.error("Error adding user:", error);
    throw error;
  }
}

export async function deleteUser(user_id, source) {
  const url = `${server}users?userID=${user_id}&source=${source}`;

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

    return response.json();
  } catch (error) {
    console.error("Error deleting user:", error);
    throw error;
  }
}

export async function updateUser(user) {
  const url = `${server}users?user_id=${user.id}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "PUT",
      headers,
      body: JSON.stringify(user),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
  } catch (error) {
    console.error("Error updating user:", error);
    throw error;
  }
}
