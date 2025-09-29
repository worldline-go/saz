<script lang="ts">
  import active from "svelte-spa-router/active";
  import { storeNoteIds } from "@/store/store";
  import { Plus } from "@lucide/svelte";
  import { push } from "svelte-spa-router";
  import { ulid } from "ulid";
</script>

<div class="bg-gray-100 border-r border-black">
  <div class="sticky top-0 overflow-auto max-h-svh no-scrollbar">
    <div class="border-b border-black leading-8">
      <a
        href="#/"
        class="block h-full hover:bg-yellow-200 hover:text-black"
        use:active={{
          path: `/`,
          className: "bg-black text-white",
          inactiveClassName: "bg-white text-black",
        }}
      >
        <span class="block px-2">SAZ</span>
      </a>
    </div>
    <div
      class="border-b border-black h-7 pl-2 flex justify-between items-center"
    >
      <span>Notes</span>
      <button
        class="text-black px-2 h-full hover:cursor-pointer hover:bg-blue-500 hover:text-white"
        onclick={() => {
          let id = ulid();

          // redirect to the new note
          push(`/note/${id}`);
        }}
      >
        <Plus />
      </button>
    </div>
    {#each $storeNoteIds as note}
      <div class="border-b border-black h-7">
        <a
          href="#/note/{note.id}"
          class="block h-full hover:bg-yellow-200 hover:text-black"
          use:active={{
            path: `/note/${note.id}`,
            className: "bg-black text-white",
            inactiveClassName: "bg-white text-black",
          }}
        >
          <span
            class="block px-2 overflow-ellipsis overflow-x-hidden whitespace-nowrap"
            title={note.name}
          >
            {note.name}
          </span>
        </a>
      </div>
    {/each}
  </div>
</div>

<style>
  @reference "tailwindcss";

  :global(.sb-link-active) {
    @apply bg-black text-white;
  }

  :global(.sb-link-inactive) {
    @apply bg-white text-black;
  }

  /* Hide scrollbar for Chrome, Safari and Opera */
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }

  /* Hide scrollbar for IE, Edge and Firefox */
  .no-scrollbar {
    -ms-overflow-style: none; /* IE and Edge */
    scrollbar-width: none; /* Firefox */
  }
</style>
