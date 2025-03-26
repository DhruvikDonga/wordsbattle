import { AppBar, Typography } from "@suid/material";
import { onMount } from "solid-js";

const Footer = () => {
    let footerRef: HTMLDivElement | null = null;
    onMount(() => {
        // Footer fade-in animation
        if (footerRef) {
          scrambleText(footerRef, "Powered by simplysocket-v0.1.0");
        }
      });

      const scrambleText = (element: HTMLDivElement, finalText: string) => {
        const chars = "!<>-_\\/[]{}—=+*^?#________"; // Random scramble characters
        let iterations = 0;
        const scrambleInterval = setInterval(() => {
          element.innerText = finalText
            .split("")
            .map((char, i) => {
              if (i < iterations) {
                return finalText[i]; // Reveal correct letter
              }
              return chars[Math.floor(Math.random() * chars.length)]; // Scramble effect
            })
            .join("");
    
          if (iterations >= finalText.length) {
            clearInterval(scrambleInterval);
            // ✅ Replace text with a clickable link after animation finishes
            element.innerHTML = `Powered by <a href="https://github.com/DhruvikDonga/simplysocket" target="_blank" style="color: #00BFA6; text-decoration: none; font-weight: bold;">simplysocket-v0.1.0</a>`;
          }
    
          iterations += 1 / 2; // Adjust speed
        }, 50);
      };
    
    return (
        <div>
            <Typography
                variant="body2"
                sx={{
                    position: "absolute",
                    bottom: "35px",
                    textAlign: "center",
                    width: "100%",
                    fontSize: "0.9rem",
                    opacity: 0.4,
                }}
                >
          Made with 🫶 by  <a href="https://dhruvikdonga.github.io" target="_blank" style="color: #00BFA6; text-decoration: none; font-weight: bold;">Dhruv!k </a>
        </Typography>
        <Typography
          //ref={(el) => (footerRef = el as HTMLDivElement)}
          variant="body2"
          sx={{
            position: "absolute",
            bottom: "10px",
            textAlign: "center",
            width: "100%",
            fontSize: "0.9rem",
            opacity: 0.4,
          }}
        >Powered by <a href="https://github.com/DhruvikDonga/simplysocket" target="_blank" style="color: #00BFA6; text-decoration: none; font-weight: bold;">simplysocket-v0.1.0</a>
        </Typography>
        </div>
    )
}

export default Footer;
