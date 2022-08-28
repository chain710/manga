<template>
  <v-dialog v-model="modalFileBrowser" max-width="450" scrollable>
    <v-card>
      <v-card-title>{{ dialogTitle }}</v-card-title>

      <v-card-text style="height: 450px">
        <v-text-field v-model="selectedPath" readonly />

        <v-list elevation="3" dense>
          <template v-if="directoryListing.hasOwnProperty('parent')">
            <v-list-item @click.prevent="select(directoryListing.parent)">
              <v-list-item-icon>
                <v-icon>mdi-arrow-left</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>{{ $t("dialog.file_browser.parent_directory") }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-divider />
          </template>

          <div v-for="(d, index) in directoryListing.entries" :key="index">
            <v-list-item @click.prevent="select(d.path)">
              <v-list-item-icon>
                <v-icon>{{ d.type === "directory" ? "mdi-folder" : "mdi-file" }}</v-icon>
              </v-list-item-icon>

              <v-list-item-content>
                <v-list-item-title>
                  {{ d.name }}
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>

            <v-divider v-if="index !== directoryListing.entries.length - 1" />
          </div>
        </v-list>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn text @click="dialogCancel">{{ $t("dialog.file_browser.cancel") }}</v-btn>
        <v-btn color="primary" @click="dialogConfirm" :disabled="!selectedPath">{{ confirmText }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  data: () => {
    return {
      directoryListing: {},
      selectedPath: "",
      modalFileBrowser: false,
    };
  },
  watch: {
    value(val) {
      if (val) {
        this.dialogInit();
      }
      this.modalFileBrowser = val;
    },
    modalFileBrowser(val) {
      !val && this.dialogCancel();
    },
  },
  props: {
    value: Boolean,
    path: {
      type: String,
      required: true,
    },
    dialogTitle: {
      type: String,
    },
    confirmText: {
      type: String,
    },
  },
  methods: {
    dialogInit() {
      try {
        this.getDirs(this.path);
        this.selectedPath = this.path;
      } catch (e) {
        this.getDirs();
      }
    },
    dialogCancel() {
      this.$emit("input", false);
    },
    dialogConfirm() {
      this.$emit("input", false);
      this.$emit("update:path", this.selectedPath);
    },
    async getDirs(path) {
      try {
        let resp = await this.$service.fsListDirectory(path);
        this.directoryListing = resp.data.data;
      } catch (e) {
        console.error(`list directory error`, e);
        this.$nerror("list_directory", e);
      }
    },
    select(path) {
      this.selectedPath = path;
      this.getDirs(path);
    },
  },
};
</script>

<style scoped></style>
