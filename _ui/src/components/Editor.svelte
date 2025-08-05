<script lang="ts">
  import CodeMirror from "svelte-codemirror-editor";
  import { sql } from "@codemirror/lang-sql";

  export let value = "";
</script>

<CodeMirror bind:value lang={sql()} />
