<!-- Copyright (c) 2024 Seagate Technology LLC and/or its Affiliates -->
<template>
  <v-container>
    <h2>CXL Agent Help Center</h2>
    <br />
    <v-row>
      <v-col v-for="(variant, i) in variants" :key="i" cols="6" md="3">
        <v-card
          class="mx-auto custom-size"
          max-width="344"
          :href="variant.url"
          target="_blank"
          rel="noopener"
          variant="tonal"
        >
          <v-card-title>
            <v-icon left>{{ variant.icon }}</v-icon>
            {{ variant.title }}
          </v-card-title>
          <v-card-text>{{ variant.content }}</v-card-text>
          <template v-slot:append>
            <v-icon icon="mdi-open-in-new"></v-icon>
          </template>
        </v-card>
      </v-col>
    </v-row>
    <br />

    <h3>Contact Support</h3>
    <p>
      If you need further assistance, please contact your local Seagate
      representative or our support team.
    </p>
    <v-icon color="#1428A0">mdi-email-outline</v-icon>
    Email:<br>
    <a href="mailto:handeok.lee@samsung.com">handeok.lee@samsung.com</a><br>
    <a href="mailto:youhyu36.lee@samsung.com">youhyu36.lee@samsung.com</a><br>
    <a href="mailto:jehoon.park@samsung.com">jehoon.park@samsung.com</a>

  </v-container>
</template>

<style scoped>
.custom-size {
  height: 150px;
}
</style>

<script>
export default {
  data() {
    return {
      variants: [
        {
          id: 0,
          title: "Pangaea GitHub Repository",
          content: "For more details, please visit our GitHub repository.",
          url: "https://github.com/SamsungDS/Pangaea",
          icon: "mdi-github",
        },
        {
          id: 1,
          title: "Pangaea Wiki",
          content:
            "For instructions on our product, please refer to the Pangaea Wiki.",
          url: "https://github.com/SamsungDS/Pangaea/wiki",
          icon: "mdi-book-open-page-variant",
        },
        {
          id: 2,
          title: "Getting Started",
          content:
            "For more details about how to deploying Pangaea, please see our Getting Started guide.",
          url: "https://github.com/SamsungDS/Pangaea/wiki/2.-Getting-Started",
          icon: "mdi-book-open-page-variant",
        },
      ],
    };
  },
};
</script>
