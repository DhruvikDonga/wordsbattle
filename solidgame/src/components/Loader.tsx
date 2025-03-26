import { createSignal, onCleanup, onMount } from "solid-js";
import gsap from "gsap";
import "@fontsource/caveat"; 

const emojiPairs = [
  ["⚔️", "🔥", "🏆"], // Battle sword + Fire = Trophy
  ["🎮", "⭐", "🏅"], // Gaming + Star = Medal
  ["💣", "⚡", "💥"], // Bomb + Lightning = Explosion
  ["🛡️", "⚒️", "👑"], // Shield + Hammer = Crown
  ["🗡️", "🛡️", "🏆"], // Sword + Shield = Trophy
];

const Loader = (props: { onComplete: () => void }) => {
  const [currentFormula, setCurrentFormula] = createSignal([...emojiPairs[0]]); // Ensure new reference
  let loaderRef: HTMLDivElement | null = null;
  let emojiInterval: number | null = null;

  onMount(() => {
    // Change emoji formula every second
    emojiInterval = window.setInterval(() => {
      const randomFormula = [...emojiPairs[Math.floor(Math.random() * emojiPairs.length)]]; // Clone array
      setCurrentFormula(randomFormula);
    }, 400);

    // GSAP Animation: Scale & Fade Loop
    gsap.fromTo(
      loaderRef,
      { opacity: 0, scale: 0.9 },
      { opacity: 1, scale: 1, duration: 0.6, repeat: -1, yoyo: true, ease: "power1.inOut" }
    );
  });

  onCleanup(() => {
    if (emojiInterval !== null) clearInterval(emojiInterval);
    gsap.killTweensOf(loaderRef); // Cleanup animations
  });

  return (
    <div
      ref={(el) => (loaderRef = el)}
      style={{
        position: "absolute",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        "font-size": "30px",
        "font-weight": "bold",
        color: "#ffffff",
        display: "flex",
        "flex-direction": "column",
        "align-items": "center",
        "text-align": "center",
        gap: "15px",
      }}
    >
      <div style={{ display: "flex", gap: "10px" }}>
        {currentFormula().map((emoji) => (
          <span style={{ "font-size": "30px" }}>{emoji}</span>
        ))}
      </div>
      <div 
        style={{ "font-size": "30px", "opacity": 0.4, "font-family": "Caveat" }} 
        class="loading-text"
      >
        Loading...
      </div>
    </div>
  );
};

export default Loader;
