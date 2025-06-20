const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

import { format } from "date-fns";

export async function getAvailabilities(user, date, duration) {
  const formattedDate = format(new Date(date), "yyyy-MM-dd");
  const url = `${server}availability?user=${user}&date=${formattedDate}&resource=${duration}`;
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

export async function fetchAppointments() {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + "appointments", {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching roles:", error);
    throw error;
  }
}

export async function fetchAppointmentsByUserID(userID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + `appointments?userID=${userID}`, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching roles:", error);
    throw error;
  }
}

export async function fetchAppointmentsBySchoolID(schoolID) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + `appointments?schoolID=${schoolID}`, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error("Error fetching roles:", error);
    throw error;
  }
}

export async function insertAppointment(appointment) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + "appointments", {
      method: "POST",
      headers,
      body: JSON.stringify(appointment),
    });
    if (!response.ok) {
      if (response.status === 409) {
        const error = new Error("unconfirmed");
        error.status = 409;
        throw error;
      }
      throw new Error("Network response was not ok");
    }
    return response.json();

  } catch (error) {
    if (error === "Unconfirmed appointments exist for this user") {
      error.message = "unconfirmed";
      throw error;
    }
    throw error;
  }
}

export async function deleteAppointment(appointment_id) {
  const url = `${server}appointments?appointmentID=${appointment_id}`;
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
    console.error("Error deleting appointment:", error);
    throw error;
  }
}

export async function updateAppointment(appointment) {
  const url = `${server}appointments?appointmentID=${appointment.id}`;
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "PUT",
      headers,
      body: JSON.stringify(appointment),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    return response.json();
  } catch (error) {
    console.error("Error updating role:", error);
    throw error;
  }
}

export async function fetchAppointmentsByToken(token) {
  const url = `${server}validate-appointment?token=${token}`;
  try {
    if (!token) {
      throw new Error("No token provided");
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
    console.error("Failed to fetch appointment details:", error);
    throw error;
  }
}

export async function confirmAppointment(token) {
  const url = `${server}validate-appointment?token=${token}`;

  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "POST",
      headers,
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    return response.json();
  } catch (error) {
    console.error("Error confirming appointment:", error);
    throw error;
  }
}

export async function cancelAppointment(token) {
  const url = `${server}validate-appointment?token=${token}`;

  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(url, {
      method: "DELETE",
      headers,
    });

    return response.json();
  } catch (error) {
    console.error("Error canceling appointment:", error);
    throw error;
  }
}
