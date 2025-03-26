import { useNavigate } from "@solidjs/router";
import { onMount } from "solid-js";
import gsap from "gsap";
import { Button, Container, Typography } from "@suid/material";
import "@fontsource/mouse-memoirs"; 
import "@fontsource/gloria-hallelujah"; 
import "@fontsource/oswald"; 
export default function NotFound() {
  const navigate = useNavigate();

  onMount(() => {
    gsap.from(".not-found", { opacity: 0, y: -20, duration: 1, ease: "power2.out" });
  });

  return (
    <Container class="not-found" sx={{ textAlign: "center", mt: 10 }}>
      <Typography variant="h2" color="error" gutterBottom fontFamily="Mouse Memoirs">
        404
      </Typography>
      <Typography variant="h5" gutterBottom fontFamily="Oswald"> 
        Oops! The page you're looking for doesn't exist.
      </Typography>
      <Button 
        variant="contained" 
        color="primary" 
        onClick={() => navigate("/")} 
        sx={{ mt: 2 }}
      >
        Go Home
      </Button>
    </Container>
  );
}
