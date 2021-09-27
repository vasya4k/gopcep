import axios from "axios";

export const HTTP = axios.create({
  baseURL: `https://127.0.0.1:1443/v1/`,
  auth: {
    username: "someuser",
    password: "somepasss"
  }
});
