import { PublicClientApplication } from "@azure/msal-browser";

const msalInstance = new PublicClientApplication({
  auth: {
    clientId: "b134585a-b99c-485a-92d7-0f331904667d",
    authority:
      "https://login.microsoftonline.com/cde3eff2-d1ad-46a3-9340-d7a530d15963",
    redirectUri: "http://localhost:8080",
  },
});

export { msalInstance };
