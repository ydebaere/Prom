const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function fetchWorkSchedules(userId) {
  try {
    const response = await fetch(`${server}workschedule?userID=${userId}`, {
      method: "GET",
      headers,
    });

    if (!response.ok) {
      throw new Error("Failed to fetch work schedules");
    }

    const data = await response.json();
    return data || [];
  } catch (error) {
    console.error("Error fetching work schedules:", error);
    throw error;
  }
}

export async function createWorkSchedule(schedule) {
  try {
    const response = await fetch(`${server}workschedule`, {
      method: "POST",
      headers,
      body: JSON.stringify(schedule),
    });

    if (!response.ok) {
      throw new Error("Failed to create work schedule");
    }

    return await response.json();
  } catch (error) {
    console.error("Error creating work schedule:", error);
    throw error;
  }
}

export async function updateWorkSchedule(schedule) {
  try {
    const response = await fetch(`${server}workschedule`, {
      method: "PUT",
      headers,
      body: JSON.stringify(schedule),
    });

    if (!response.ok) {
      throw new Error("Failed to update work schedule");
    }

    return await response.json();
  } catch (error) {
    console.error("Error updating work schedule:", error);
    throw error;
  }
}

export async function deleteWorkSchedule(scheduleId, userId) {
  try {
    const response = await fetch(`${server}workschedule`, {
      method: "DELETE",
      headers,
      body: JSON.stringify({ id: scheduleId, user_id: userId }),
    });

    if (!response.ok) {
      throw new Error("Failed to delete work schedule");
    }

    return await response.json();
  } catch (error) {
    console.error("Error deleting work schedule:", error);
    throw error;
  }
}
