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
  lineNumbers={false}
  foldGutter={false}
  highlight={{ activeLine: false, activeLineGutter: false }}
  class="{collapse ? 'h-7 overflow-visible' : 'min-h-7'} {className}"
  {readonly}
/>
