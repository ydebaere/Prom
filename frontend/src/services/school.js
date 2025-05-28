const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

export async function fetchSchools(dirID) {
  if (!token) {
    try {
      const response = await fetch(server + "schools", {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${"anonymous-fetch"}`,
        },
      });

      if (!response.ok) {
        throw new Error("Network response was not ok");
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error("Error fetching schools:", error);
      throw error;
    }
  } else {
    if (dirID) {
      try {
        const response = await fetch(server + `schools?dirID=${dirID}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Network response was not ok");
        }
        const data = await response.json();
        const schools = data.map((school) => ({
          address: school.address,
          director: school.director,
          director_name: school.director_name,
          email: school.email,
          phone: school.phone,
          id: school.id,
          name: school.name,
        }));
        return schools;
      } catch (error) {
        console.error("Error fetching schools:", error);
        throw error;
      }
    } else {
      try {
        const response = await fetch(server + "schools", {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Network response was not ok");
        }

        const data = await response.json();
        return data;
      } catch (error) {
        console.error("Error fetching schools:", error);
        throw error;
      }
    }
  }
}

export async function insertSchool(school) {
  try {
    if (!token) {
      throw new Error("No token found");
    }
    const response = await fetch(server + "schools", {
      method: "POST",
      headers,
      body: JSON.stringify(school),
    });

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  } catch (error) {
    console.error("Error adding school:", error);
    throw error;
  }
}

export async function deleteSchool(school_id, admin) {
  const url = `${server}schools?schoolID=${school_id}&admin=${admin}`;

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
    console.error("Error deleting school:", error);
    throw error;
  }
}
