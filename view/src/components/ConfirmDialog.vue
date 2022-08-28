<template>
  <v-dialog v-model="modal" max-width="450">
    <v-card>
      <v-card-title :class="titleClass">{{ title }}</v-card-title>

      <v-card-text>
        <v-container fluid>
          <v-row>
            <v-col v-if="body">
              <slot>{{ body }}</slot>
            </v-col>
          </v-row>

          <v-row v-if="confirmText">
            <v-col>
              <v-checkbox v-model="confirmation" :color="color">
                <template v-slot:label>
                  {{ confirmText }}
                </template>
              </v-checkbox>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn text @click="dialogCancel">{{ buttonCancel || $t("global.cancel") }}</v-btn>
        <v-btn :color="color" :loading="loading" @click="dialogConfirm" :disabled="confirmText && !confirmation">
          {{ buttonConfirm || $t("global.confirm") }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  data: function () {
    return {
      modal: this.value,
      confirmation: false,
      loading: false,
    };
  },
  props: {
    value: Boolean,
    title: {
      type: String,
      required: true,
    },
    body: {
      type: String,
      required: false,
    },
    confirmText: {
      type: String,
      required: false,
    },
    confirmFunc: {
      type: Function,
      required: false,
    },
    buttonCancel: {
      type: String,
      required: false,
    },
    buttonConfirm: {
      type: String,
      required: false,
    },
    type: {
      type: String,
      default: "primary",
    },
  },
  watch: {
    value(val) {
      this.modal = val;
    },
    modal(val) {
      !val && this.dialogCancel();
    },
  },
  methods: {
    dialogCancel() {
      this.confirmation = false;
      this.$emit("input", false);
    },
    async dialogConfirm() {
      if (!this.confirmFunc) {
        this.$emit("confirm");
        this.$emit("input", false);
      }

      try {
        this.loading = true;
        await this.confirmFunc();
      } catch (error) {
        console.log(`confirm func error`, error);
      } finally {
        this.loading = false;
        this.$emit("confirm");
        this.$emit("input", false);
      }
    },
  },
  computed: {
    color() {
      return this.type;
    },
    titleClass() {
      switch (this.type) {
        case "error":
          return "red white--text";
        default:
          return this.type;
      }
    },
  },
};
</script>

<style scoped></style>
