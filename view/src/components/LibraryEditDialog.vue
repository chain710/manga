<template>
  <v-dialog
    v-model="modal"
    :fullscreen="this.$vuetify.breakpoint.xsOnly"
    :hide-overlay="this.$vuetify.breakpoint.xsOnly"
    max-width="700"
    scrollable>
    <form novalidate>
      <v-card>
        <v-toolbar class="hidden-sm-and-up">
          <v-btn icon @click="dialogClose">
            <v-icon>mdi-close</v-icon>
          </v-btn>
          <v-toolbar-title>{{ dialogTitle }}</v-toolbar-title>
          <v-spacer />
          <v-toolbar-items>
            <v-btn text color="primary" @click="dialogConfirm">{{ confirmText }}</v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <v-card-title class="hidden-xs-only">{{ dialogTitle }}</v-card-title>

        <v-card-text class="pa-0">
          <v-tabs :vertical="$vuetify.breakpoint.smAndUp" v-model="tab">
            <v-tab class="justify-start">
              <v-icon left class="hidden-xs-only">mdi-bookshelf</v-icon>
              {{ $t("dialog.edit_library.tab_general") }}
            </v-tab>

            <v-tab-item>
              <v-card flat :min-height="minHeight">
                <v-container fluid>
                  <!--name-->
                  <v-row>
                    <v-col>
                      <v-text-field
                        v-model="form.name"
                        autofocus
                        :label="$t('dialog.edit_library.field_name')"
                        :error-messages="getErrors('name')"
                        @input="v$.form.name.$touch()"
                        @blur="v$.form.name.$touch()" />
                    </v-col>
                  </v-row>

                  <v-row justify="center">
                    <v-col cols="8" align-self="center">
                      <file-browser-dialog
                        v-model="showFileBrowser"
                        :path.sync="form.path"
                        :confirm-text="$t('dialog.file_browser.confirm')"
                        :dialog-title="$t('dialog.file_browser.title')" />

                      <v-text-field
                        v-model="form.path"
                        :label="$t('dialog.edit_library.path')"
                        :error-messages="getErrors('path')"
                        @input="v$.form.path.$touch()"
                        @blur="v$.form.path.$touch()" />
                    </v-col>
                    <v-col cols="4" align-self="center">
                      <v-btn @click="showFileBrowser = true">{{ $t("dialog.edit_library.button_browse") }}</v-btn>
                    </v-col>
                  </v-row>
                </v-container>
              </v-card>
            </v-tab-item>
          </v-tabs>
        </v-card-text>
        <!--buttons-->
        <v-card-actions class="hidden-xs-only">
          <v-spacer />
          <v-btn text @click="dialogClose">{{ $t("dialog.edit_library.button_cancel") }}</v-btn>
          <v-btn color="primary" :loading="loading" @click="dialogConfirm">{{ confirmText }}</v-btn>
        </v-card-actions>
      </v-card>
    </form>
  </v-dialog>
</template>
<script>
import useVuelidate from "@vuelidate/core";
import FileBrowserDialog from "./FileBrowserDialog.vue";
import { required, maxLength } from "./util/i18n-validators";
export default {
  setup() {
    return {
      v$: useVuelidate(),
    };
  },
  components: { FileBrowserDialog },
  data: function () {
    return {
      modal: this.value,
      showFileBrowser: false,
      tab: 0,
      loading: false,
      form: {
        name: "",
        path: "/",
      },
    };
  },
  props: {
    value: Boolean,
    library: Object,
  },
  validations: {
    form: {
      name: { required, maxLength: maxLength(32) },
      path: { required },
    },
  },
  methods: {
    dialogClose() {
      this.$emit("input", false);
      this.tab = 0;
    },
    dialogConfirm() {
      this.editLibrary();
    },
    getErrors(fieldName) {
      let errors = [];
      const field = this.v$.form[fieldName];
      if (field && field.$invalid && field.$dirty && field.$errors) {
        for (const err of field.$errors) {
          errors.push(err.$message);
        }
      }

      return errors;
    },
    validateLibrary() {
      this.v$.$touch();

      if (this.v$.$invalid) {
        return null;
      }
      let lib = {
        name: this.form.name,
        path: this.form.path,
      };
      if (this.library) {
        lib["id"] = this.library.id;
      }
      return lib;
    },
    async editLibrary() {
      const library = this.validateLibrary();
      if (!library) {
        this.tab = 0;
        return;
      }

      this.loading = true;
      if (library.id) {
        await this.patch(library);
      } else {
        await this.add(library);
      }
      this.loading = false;
    },
    async add(library) {
      try {
        const resp = await this.$service.addLibrary(library);
        console.log(`add library ok`, resp);
        this.dialogClose();
        this.$ninfo(`add_library`);
        this.$emit(`updated`);
      } catch (error) {
        this.$nerror(`add_library`, error);
        console.log(`add library error`, error);
      }
    },
    async patch(library) {
      try {
        const resp = await this.$service.patchLibrary(library.id, library);
        console.log(`patch library ok`, resp);
        this.dialogClose();
        this.$ninfo(`patch_library`);
        this.$emit(`updated`);
      } catch (error) {
        this.$nerror(`patch_library`, error);
        console.log(`patch library error`, error);
      }
    },

    resetDialog() {
      const lib = this.library;
      this.form.name = lib ? lib.name : "";
      this.form.path = lib ? lib.path : "/";
    },
  },
  computed: {
    dialogTitle() {
      return this.library ? this.$t("dialog.edit_library.edit_title") : this.$t("dialog.edit_library.add_title");
    },
    confirmText() {
      if (this.library) {
        return this.$t("dialog.edit_library.button_edit");
      } else {
        return this.$t("dialog.edit_library.button_add");
      }
    },
    minHeight() {
      return this.$vuetify.breakpoint.xs ? this.$vuetify.breakpoint.height * 0.8 : undefined;
    },
  },
  watch: {
    value(val) {
      this.modal = val;
    },
    modal(val) {
      if (val) {
        this.resetDialog();
      } else {
        this.dialogClose();
      }
    },
  },
};
</script>
