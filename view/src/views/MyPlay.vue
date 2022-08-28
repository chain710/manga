<!-- <template>
  <div class="ma-0 pa-0">
    <dragable-resizable :w="1000" :h="300" lock-aspect-ratio :resizable="false" :z="1" drag-handle=".drag-handle">
      <div class="drag-handle" style="background-color: aqua; width: 100%; height: 100%"></div>
      <dragable-resizable
        :w="100"
        :h="200"
        lock-aspect-ratio
        :z="2"
        parent
        drag-handler=".drag-inner"
        style="top: 0px; left: 0px">
        <div class="drag-inner" style="background-color: red; width: 100%; height: 100%">
          <span>inner</span>
        </div>
      </dragable-resizable>
    </dragable-resizable>
  </div>
</template> -->
<template>
  <div style="position: absolute; background-color:gray; width: 200px; height: 200px;">
    <div class="menu-icon">
      <v-menu bottom left>
        <template v-slot:activator="{ on, attrs }">
          <v-btn dark icon v-bind="attrs" v-on="on">
            <v-icon>mdi-dots-vertical</v-icon>
          </v-btn>
        </template>

        <v-list dense>
          <v-list-item>
            <v-list-item-title>tt1</v-list-item-title>
          </v-list-item>
          <v-list-item>
            <v-list-item-title>tt2</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </div>
  </div>
</template>
<script>
export default {
  data: () => ({
    items: [{ title: "Click Me" }, { title: "Click Me" }, { title: "Click Me" }, { title: "Click Me 2" }],
  }),
};

// export default {
//   mounted() {
//     console.log(`window mounted ${window.innerHeight} ${window.innerWidth}`);
//   },
//   methods: {
//     onresize() {
//       console.log(`window resize ${window.innerHeight} ${window.innerWidth}`);
//     }
//   }
// }
</script>
<style scoped>
.menu-icon {
  position: absolute;
  bottom: 5px;
  right: 5px;
}
</style>
