import { createSignal, createEffect, onCleanup, For } from "solid-js";
import { useNavigate } from "@solidjs/router";
import gsap from "gsap";
import Button from "@suid/material/Button";
import Container from "@suid/material/Container";
import Card from "@suid/material/Card";
import CardContent from "@suid/material/CardContent";
import Typography from "@suid/material/Typography";
import ThemeProvider from "@suid/material/styles/ThemeProvider";
import createTheme from "@suid/material/styles/createTheme";
import useMediaQuery from "@suid/material/useMediaQuery";
import "@fontsource/mouse-memoirs"; 
import "@fontsource/gloria-hallelujah"; 
import "@fontsource/oswald"; 
import Footer from "./Footer";
import Loader from "./Loader";
import { A } from "@solidjs/router";

// 🌙 Dark Theme
const darkTheme = createTheme({
  palette: {
    mode: "dark",
    background: { default: "#1E1E1E", paper: "#252525" },
    primary: { main: "#007BFF" },
  },
});

const gameCards = [
  { title: "Clash of Words", intro: "Challenge your opponent to test your vocabulary." },
];

const GameSection = () => {
  const [currentIndex, setCurrentIndex] = createSignal(0);
  let cardContainerRef: HTMLDivElement | null = null;
  let curvyLineRef: SVGPathElement | null = null;
  const isSmallScreen = useMediaQuery("(max-width: 768px)");
  const [loading, setLoading] = createSignal(true);
  
  const navigate = useNavigate();

  createEffect(() => {
    setTimeout(() => setLoading(false), 2000);
  });

  const nextCard = () => setCurrentIndex((prev) => (prev + 1) % gameCards.length);
  const prevCard = () => setCurrentIndex((prev) => (prev - 1 + gameCards.length) % gameCards.length);

  // Slide animation effect
  createEffect(() => {
    if (!loading()) {

    if (cardContainerRef) {
      gsap.to(cardContainerRef, {
        x: -currentIndex() * 80 + "vw",
        duration: 0.8,
        ease: "power3.out",
      });
    }
  }
  });

  // Curvy Line Animation
  createEffect(() => {
    if (!loading()) {

    if (curvyLineRef) {
      gsap.to(curvyLineRef, {
        strokeDasharray: "300 50",
        strokeDashoffset: -350,
        duration: 2.5,
        repeat: -1,
        ease: "linear",
      });
    }
  }
  });

  onCleanup(() => {
    gsap.killTweensOf(cardContainerRef);
    gsap.killTweensOf(curvyLineRef);
  });

  const makeroom = (length:number) => {
    let result = "";
    const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
  };

  const playWithFriend = () => {
    const roomname = makeroom(10);
    navigate(`/clashofwords/play?room=${roomname}`);
  };

  const playWithRandomFriend = () => {
    navigate("/clashofwords/play-random");
  };

  return (
    <ThemeProvider theme={darkTheme}>
    {loading() ? <Loader onComplete={() => setLoading(false)} /> : 
      <Container
        disableGutters
        sx={{
          height: "90vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: darkTheme.palette.background.default,
          color: "#fff",
          position: "relative",
          overflow: "hidden",
          width: isSmallScreen() ? "95vw" : "80vw",
        }}
      >
        {gameCards.length > 1 && (
          <Button
            variant="contained"
            sx={{ position: "absolute", left: "10px", zIndex: 100, minWidth: "30px" }}
            onClick={prevCard}
          >
            ◀
          </Button>
        )}

        <div
          ref={(el) => (cardContainerRef = el)}
          style={{ 
            display: "flex", 
            width: gameCards.length > 1 ? `${gameCards.length * 15}vw` : "80vw",  
            transition: "transform 0.5s ease-out", 
            "justify-content": gameCards.length === 1 ? "center" : "flex-start"
          }}
        >
          <For each={gameCards}>{(card, i) => (
            <div style={{ flex: "0 0 80vw", display: "flex", "justify-content": "center", "align-items": "center" }}>
              <Card
                sx={{
                  width: isSmallScreen() ? "80vw" : "60vw",
                  height: isSmallScreen() ? "60vh" : "70vh",
                  padding: "24px",
                  textAlign: "center",
                  background: "#252525",
                  display: "flex",
                  flexDirection: "column",
                  justifyContent: "center",
                  alignItems: "center",
                  transform: `scale(${currentIndex() === i() ? 1 : 0.8})`,
                  opacity: currentIndex() === i() ? 1 : 0.5,
                  position: "relative",
                  overflow: "hidden",
                }}
              >


                <CardContent sx={{ display: "flex", flexDirection: "column", alignItems: "center", position: "relative", zIndex: 2 }}>
                  <Typography variant="h1" sx={{ marginBottom: "15px" }} fontFamily="Mouse Memoirs" >{card.title}</Typography>
                  <Typography variant="h5" sx={{ marginBottom: "20px" }} fontFamily="Gloria Hallelujah">{card.intro}</Typography>
                  <div style={{ display: "flex", "justify-content": "center", gap: "10px" }}>
                    <Button variant="outlined" sx={{ backgroundColor: "#007BFF", color: "#fff"  }}  onclick={playWithFriend}>🎮 &nbsp;<span style={{"font-family":"Oswald"}}>Play with F.r.i.e.n.d.s</span></Button>
                    <Button variant="outlined" sx={{ backgroundColor: "#007BFF", color: "#fff" }}  onclick={playWithRandomFriend}>🔀 &nbsp; <span style={{"font-family":"Oswald"}}>Random Game</span></Button>
                  </div>
                </CardContent>
                {/* Curvy Line Animation */}
                <svg
                  width="100%"
                  height="100%"
                  viewBox={isSmallScreen() ?  "0 0 140 20" : "0 0 300 100"}
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  style={{
                    position: "absolute",
                    top: 0,
                    left: 0,
                    opacity: 0.2,
                  }}
                >
                  <path
                    ref={(el) => (curvyLineRef = el)}
                    d="M0 100 C100 0, 200 200, 300 100 S400 200, 500 100"
                    stroke="#00BFA6"
                    stroke-width="3"
                    fill="transparent"
                    stroke-linecap="round"
                    id="curvyLine"
                  />
                </svg>
              </Card>
            </div>
          )}</For>
        </div>

        {gameCards.length > 1 && (
          <Button
            variant="contained"
            sx={{ position: "absolute", right: "10px", zIndex: 100, minWidth: "30px" }}
            onClick={nextCard}
          >
            ▶
          </Button>
        )}
      </Container>
    }
      <Footer />

    </ThemeProvider>
  );
};

export default GameSection;
