import AppBar from "@suid/material/AppBar";
import Toolbar from "@suid/material/Toolbar";
import { onMount } from "solid-js";
import gsap from "gsap";
import { TextPlugin } from "gsap/TextPlugin";

import "@fontsource/pacifico"; // Import Pacifico font for cursive style

const Header = () => {
  let textRef: SVGTextElement | null = null;
  gsap.registerPlugin(TextPlugin);

  onMount(() => {
    if (textRef) {
      gsap.fromTo(
        textRef,
        { strokeDasharray: 500, strokeDashoffset: 500, fillOpacity: 0 },
        {
          strokeDashoffset: 0,
          duration: 10, // Faster handwriting effect
          ease: "power2.out",
        }
      );
  
      // Start fading fill midway instead of waiting for stroke to finish
      gsap.to(textRef, {
        fillOpacity: 1,
        stroke: "none",
        duration: 1.5, // Faster transition
        delay: 2.5, // Start filling before stroke ends
      });
    }
  });

  return (
    <AppBar position="static" sx={{ backgroundColor: "#0e03a3", padding: "1px 0" }}>
      <Toolbar sx={{ justifyContent: "center" }}>
        <svg width="300" height="70" viewBox="0 0 300 80">
          <text
            ref={(el) => (textRef = el)}
            x="50%"
            y="50%"
            text-anchor="middle"
            dy=".3em"
            font-size="40"
            font-family="'Pacifico', cursive"
            stroke="white"
            fill="white"
            fill-opacity={0}
            stroke-width={2}
            stroke-linecap="round"
          >
            Words Battle
          </text>
        </svg>
      </Toolbar>
    </AppBar>
  );
};

export default Header;
