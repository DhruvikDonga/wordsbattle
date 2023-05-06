<template>
  <v-container>
    <v-row class="text-center">
      <v-col cols="12">
        <v-img
          :src="require('../assets/logo.svg')"
          class="my-3"
          contain
          height="200"
        />
      </v-col>

      <v-col class="mb-4">
        <h1 class="display-2 font-weight-bold mb-3">
          Welcome to the Wordbattles Beta
        </h1>


        <p class="subheading font-weight-regular">
            
            <p>Current online :- {{ active }}</p>
          
          Let the battle begin 
          </p>
          <v-container>
    <v-row align="center" justify="center">
     
    <v-col cols="auto">
    <v-btn block  rounded>Random Game</v-btn>
    </v-col>
    <v-col cols="auto">
          <v-btn block rounded>Play with a friend</v-btn>
      </v-col>
    </v-row>
    </v-container>
    </v-col>
    </v-row>
  </v-container>
</template>


<script>
export default {
  name: 'HelloWorld',

  data() {
    return {
      
      ws: null,
      active:null,
    }
    }
}
</script>
