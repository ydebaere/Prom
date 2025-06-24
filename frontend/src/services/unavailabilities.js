const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function getAvailabilities(user_id, date) {
  const formattedDate = format(new Date(date), "yyyy-MM-dd");
  const url = `${server}availability?userID=${user_id}&date=${formattedDate}`;
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
    return data;
  } catch (error) {
    console.error("Error fetching slots:", error);
    throw error;
  }
}

export async function fetchUnavailabilities() {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(`${server}unavailabilities`, {
      method: "GET",
      headers,
    });
    if (!response.ok) {
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching unavailabilities:", error);
    throw error;
  }
}

export async function fetchUserUnavailabilities(userID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(`${server}unavailabilities?userID=${userID}`, {
      method: "GET",
      headers,
    });
    if (!response.ok) {
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching unavailabilities:", error);
    throw error;
  }
}

export async function insertUnavailability(data) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(`${server}unavailabilities`, {
      method: "POST",
      headers,
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  } catch (error) {
    console.error("Error adding unavailability:", error);
    throw error;
  }
}

export async function deleteUnavailability(id) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(
      `${server}unavailabilities?unavailabilityID=${id}`,
      {
        method: "DELETE",
        headers,
      }
    );

    if (!response) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  } catch (error) {
    console.error("Error deleting unavailability:", error);
    throw error;
  }
}

export async function updateUnavailability(data) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(
      `
    ${server}unavailabilities`,
      {
        method: "PUT",
        headers,
        body: JSON.stringify(data),
      }
    );
    if (!reponse) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  } catch (error) {
    console.error("Error updating unavailability:", error);
    throw error;
  }
}
