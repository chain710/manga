<template>
  <div>
    <v-app-bar app color="primary" dark>
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
    </v-app-bar>
    <v-navigation-drawer app v-model="drawer">
      <v-list-item
        @click="$router.push({ name: 'home' })"
        inactive
        class="pb-2"
      >
        <v-list-item-avatar>
          <v-img src="../assets/logo.svg" />
        </v-list-item-avatar>

        <v-list-item-content>
          <v-list-item-title class="title"> Manga </v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-divider></v-divider>

      <v-list dense nav>
        <v-list-item v-for="item in items" :key="item.title" :to="item.to">
          <v-list-item-icon>
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
    <v-main>
      <router-view></router-view>
    </v-main>
  </div>
</template>
<script>
export default {
  data: () => ({
    drawer: true,
    items: [
      { title: "Dashboard", icon: "mdi-view-dashboard", to: "/" },
      { title: "Photos", icon: "mdi-image", to: "/test" },
      { title: "About", icon: "mdi-help-box", to: "/b" },
    ],
  }),

  created: function () {
    console.log("main created");
  },
  mounted: function () {
    console.log("main mounted");
  },
};
</script>
