import { onMount } from "solid-js";
import { Button, Card, CardContent, Container, TextField, Typography, useMediaQuery } from "@suid/material";
import gsap from "gsap";

export default function PlayerDetails(props: any) {
  let elementRef: any;

  onMount(() => {
    if (elementRef) {
      gsap.fromTo(
        elementRef,
        { opacity: 0, y: 20 },
        { opacity: 1, y: 0, duration: 0.8, ease: "power2.out" }
      );
    }
  });
    const isSmallScreen = useMediaQuery("(max-width: 768px)");
  

  return (
    <div ref={(el) => (elementRef = el)} class="flex items-center justify-center min-h-screen">
      <Container class="flex justify-center"
        disableGutters
        sx={{
          height: "90vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
          overflow: "hidden",
          width: isSmallScreen() ? "95vw" : "80vw",
        }}
      >
        <Card class="mx-auto w-full max-w-sm p-4" variant="outlined">
          <CardContent>
            <div class="text-center" style={{"justify-content": "center"}}>
              <Typography variant="h2" sx={{ marginBottom: "15px" }} fontFamily="Mouse Memoirs" >Care to enter your name</Typography>
              <Typography variant="body1" sx={{ marginBottom: "20px" }} fontFamily="Gloria Hallelujah">This is only for your friend to recognize you, but you can skip it.</Typography>
              <br />
                <form ref={props.formRef} onSubmit={props.sendNewName}>
                <div style={{"justify-content": "center", "padding-bottom": "10px"}} >
                    <TextField
                    value={props.firstName()}
                    onChange={(e) => props.setFirstName(e.target.value)}
                    label="Good Name"
                    error={!props.formValid() && props.firstName().length > 0}
                    helperText={
                        !props.formValid() && props.firstName().length > 0
                        ? "Name must be between 3 to 10 letters"
                        : ""
                    }
                    required
                    fullWidth
                    />
                    </div>
                    <div style={{ display: "flex", "justify-content": "center", gap: "10px" }}>
                    <Button variant="outlined" sx={{ backgroundColor: "#007BFF", color: "#fff"  }} type="submit">
                        Submit 👍
                    </Button>
                    <Button onClick={props.notSendNewName} variant="outlined" sx={{ backgroundColor: "#007BFF", color: "#fff"  }}>
                        No need 😁
                    </Button>
                    </div>
                </form>
              </div>
          </CardContent>
        </Card>
      </Container>
    </div>
  );
}
