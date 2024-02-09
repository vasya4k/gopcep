import axios from 'axios';

export const HTTP = axios.create({
    baseURL: `https://127.0.0.1:1443/v1/`,
    // need for UI development
    // auth: {
    //     username: 'admin',
    //     password: 'pass'
    // }
});
