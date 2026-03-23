<!-- Copyright (c) 2024 Seagate Technology LLC and/or its Affiliates -->
<template>
  <v-list nav>
    <v-list-item
      prepend-avatar="@/assets/icons/Icons-House-Web-Blue.png"
      to="/dashboard"
      title="Dashboard"
      value="Dashboard"
    ></v-list-item>

    <v-list-item
      prepend-avatar="@/assets/icons/Icons-Data-Center-Blue.png"
      to="/appliances"
      title="Fabrics & Switches"
      value="Appliances"
    ></v-list-item>

    <v-list-item
      prepend-avatar="@/assets/icons/Icons-Server-Web-blue.png"
      to="/hosts"
      title="CXL-Hosts"
      value="CXL-Hosts"
    ></v-list-item>
    <v-list-item
      prepend-avatar="@/assets/icons/Icons-Questions--Web-Blue.png"
      to="/help"
      title="Help"
      value="Help"
    ></v-list-item>
    <!--
    <v-list-item
      prepend-avatar="@/assets/icons/Icons-Lightbulb-Web-Blue.png"
      to="/alerts"
      title="Alerts"
      value="Alerts"
    ></v-list-item>
    -->
  </v-list>
</template>
