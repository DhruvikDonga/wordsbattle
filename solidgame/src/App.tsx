import { createSignal, onCleanup, onMount } from "solid-js";
import ThemeProvider from "@suid/material/styles/ThemeProvider";
import createTheme from "@suid/material/styles/createTheme";
import gsap from "gsap";
import Header from "./components/Header";
import Footer from "./components/Footer";
import GameSection from "./components/GameSection";
import Loader from "./components/Loader";
import AppRouter from "./router";

// 🌙 Dark Mode Theme
const darkTheme = createTheme({
  palette: {
    mode: "dark",
    background: { default: "#1E1E1E", paper: "#252525" },
    primary: { main: "#0e03a3" },
  },
});

const App = () => {
  let contentRef: HTMLDivElement | null = null;


  onMount(() => {
    if (contentRef) gsap.set(contentRef, { opacity: 0 }); // Initially hide the content
  });

  return (
    <ThemeProvider theme={darkTheme}>
      <div style={{ "min-height": "100vh", overflow: "hidden", position: "relative" }}>
        <Header />
        
        <AppRouter />        
      </div>
    </ThemeProvider>
  );
};

export default App;
