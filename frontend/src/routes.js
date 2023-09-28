import Login from './Login.svelte';
import Requests from './Requests.svelte';
import Register from "./Register.svelte";

 const routes = {
  '/': Login,
  '/requests': Requests,
  '/register': Register
};
export default routes;