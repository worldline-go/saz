<script lang="ts">
  import CodeMirror from "svelte-codemirror-editor";
  import { sql } from "@codemirror/lang-sql";

  let {
    value = $bindable(),
    collapse = false,
    readonly = false,
    class: className = "",
  } = $props();
</script>

<CodeMirror
  bind:value
  lang={sql()}
  class="{collapse ? 'h-7' : 'min-h-7'} {className}"
  {readonly}
/>
