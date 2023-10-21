import Login from './Login.svelte';
import Requests from './Requests.svelte';
import Register from "./Register.svelte";
import Request from "../Request.svelte";
import LandingPage from "./LandingPage.svelte";

 const routes = {
  '/': LandingPage,
  '/login': Login,
  '/requests': Requests,
  '/register': Register,
  '/request': Request
};
export default routes;