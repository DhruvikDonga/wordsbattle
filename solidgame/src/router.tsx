import { Router, Route, useLocation } from "@solidjs/router";
import { createEffect } from "solid-js";
import gsap from "gsap";
import PlayClashofWords from "./components/clashofwords";
import GameSection from "./components/GameSection";
import NotFound from "./components/NotFound";


const AppRouter = () => {
  //const location = useLocation(); // Watch for route changes

  return (
    <div class="page">
      <Router>
          <Route path="/" component={GameSection} />
          <Route path="/clashofwords/play" component={PlayClashofWords} />
          <Route path="/clashofwords/play-random" component={PlayClashofWords} />
          <Route path="*" component={NotFound} /> 
      </Router>
    </div>
  );
};

export default AppRouter;
