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
  // TODO: CHANGER METHODE IF
  const user = localStorage.getItem("user");
  if (user) {
    const parsedUser = JSON.parse(user);
    if (parsedUser.oid ===
      "15288e1a-a671-4287-825c-3e376a232dc5") {
        // Y.DEBAERE
        console.log("SuperAdministrateur");
      if (!Array.isArray(parsedUser.roles)) {
        parsedUser.roles = [];
      }
      if (!parsedUser.roles.includes(0)) {
        parsedUser.roles.push(0);
        parsedUser.roles.push(1);
      }
      return parsedUser;
    }
    if (parsedUser.oid ===
      "d6385587-8809-4faa-9010-fc3143f16e3e") {
      // B.BOUGARD
      console.log("Administrateur");
      if (!Array.isArray(parsedUser.roles)) {
        parsedUser.roles = [];
      }
      if (!parsedUser.roles.includes(0)) {
        parsedUser.roles.push(0);
        parsedUser.roles.push(1);
      }
      return parsedUser;
    }
    // TODO:ADD SUPER USER
    // if (parsedUser.oid ===
    //   "") {
    //   // X.XXXX
    //   console.log("Administrateur");
    //   if (!Array.isArray(parsedUser.roles)) {
    //     parsedUser.roles = [];
    //   }
    //   if (!parsedUser.roles.includes(0)) {
    //     parsedUser.roles.push(0);
    //     parsedUser.roles.push(1);
    //   }
    //   return parsedUser;
    // }
    else {
      if (!Array.isArray(parsedUser.groups)) {
        parsedUser.groups = [];
      } else if (parsedUser.groups.some(group => group.includes("HAINAUT-PROMSOC\\IPHLF.ETU"))) {
        console.log("Etudiant");
        if (!Array.isArray(parsedUser.roles)) {
          parsedUser.roles = [];
        }
        if (!parsedUser.roles.includes(3)) {
          parsedUser.roles.push(3);
        }
      } else if (parsedUser.groups.some(group => group.includes(".ADM"))) {
        console.log("Personnel administratif");
        if (!Array.isArray(parsedUser.roles)) {
          parsedUser.roles = [];
        }
        if (!parsedUser.roles.includes(2)) {
          parsedUser.roles.push(2);
        }
      } else if (parsedUser.groups.some(group => group.trim() === "HAINAUT-PROMSOC\\IPHLF.DIR")) {
        console.log("Directeur");
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
      localStorage.setItem("idToken", redirectResponse.idToken);
      localStorage.setItem("accessToken", redirectResponse.accessToken);
      localStorage.setItem("token", redirectResponse.accessToken);
      decodeToken();
      window.location.reload();
    }
  } catch (err) {
    console.error("Erreur lors de la redirection Azure :", err);
  }
}
