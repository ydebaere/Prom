const server = import.meta.env.VITE_BASE_URL;
const token = localStorage.getItem("token");
const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${token}`,
};

import { msalInstance } from "boot/msal";

export const user = JSON.parse(localStorage.getItem("user"));


export async function downloadAppointments(token, unique_name) {
  const response = await fetch(server + `download?unique_name=${unique_name}`, {
    method: "GET",
    headers,
  });
  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  return data;
}

export function getUser() {
  const user = localStorage.getItem("user");
  if (user) {
    const parsedUser = JSON.parse(user);
    if (parsedUser.oid ===
      // Y.DeBaere
      // G.Libens
      "15288e1a-a671-4287-825c-3e376a232dc5" ||
      "9c1792fe-5cee-4ac1-8f1b-1aabb8fda10f" ) {
      if (!Array.isArray(parsedUser.roles)) {
        parsedUser.roles = [];
      }
      if (!parsedUser.roles.includes(0)) {
        parsedUser.roles.push(0);
        parsedUser.roles.push(1);
      }
      return parsedUser;
    }
    else {
      if (!Array.isArray(parsedUser.groups)) {
        parsedUser.groups = [];
      } else if (parsedUser.groups.some(group => group.includes(".ETU"))) {
        console.log("Student");
        if (!Array.isArray(parsedUser.roles)) {
          parsedUser.roles = [];
        }
        if (!parsedUser.roles.includes(3)) {
          parsedUser.roles.push(3);
        }
      } else if (parsedUser.groups.some(group => group.includes(".ADM"))) {
        console.log("Secretary");
        if (!Array.isArray(parsedUser.roles)) {
          parsedUser.roles = [];
        }
        if (!parsedUser.roles.includes(2)) {
          parsedUser.roles.push(2);
        }
      } else if (parsedUser.groups.some(group => group.includes(".DIR"))) {
        console.log("Director");
        if (!Array.isArray(parsedUser.roles)) {
          parsedUser.roles = [];
        }
        if (!parsedUser.roles.includes(1)) {
          parsedUser.roles.push(1);
        }
      }
      return parsedUser;
    }
  }
  return null;
}

export function decodeToken() {
  try {
    let token = localStorage.getItem("token");
    if (!token) {
      throw new Error("No token found");
    }
    const payload = token.split(".")[1];
    const decoded = atob(payload);
    const userData = JSON.parse(decoded);

    // Ajouter l'objet user au local storage
    localStorage.setItem("user", JSON.stringify(userData));
    return userData;
  } catch (error) {
    console.error("Error decoding token:", error);
    return null;
  }
}

export function checkTokenExpiry() {
  const tokenExpiry = localStorage.getItem("tokenExpiry");
  if (tokenExpiry && Date.now() > tokenExpiry) {
    alert("Your session has expired. Please log in again.");
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    localStorage.removeItem("tokenExpiry");
    window.location.href = "/";
  }
}

export async function azureLogin() {
  try {
    await msalInstance.initialize();

    const redirectResponse = await msalInstance.handleRedirectPromise();
    if (redirectResponse) {
      localStorage.setItem("idToken", redirectResponse.account.idToken);
      localStorage.setItem("accessToken", redirectResponse.accessToken);
      localStorage.setItem("token", redirectResponse.accessToken);
      decodeToken();
      window.location.reload();
    }
  } catch (err) {
    console.error("Erreur lors de la redirection Azure :", err);
  }
}
