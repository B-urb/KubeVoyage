import Login from './Login.svelte';
import Requests from './Requests.svelte';
import Register from "./Register.svelte";
import Request from "../Request.svelte";

 const routes = {
  '/': Login,
  '/requests': Requests,
  '/register': Register,
  '/request': Request
};
export default routes;