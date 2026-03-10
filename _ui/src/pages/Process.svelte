<script lang="ts">
  import { onMount } from "svelte";
  import {
    requestProcesses,
    requestProcessTerminate,
    requestProcessDelete,
    type ProcessQueryParams,
  } from "@/helper/call";
  import { addToast } from "@/store/toast";
  import {
    RefreshCw,
    Trash,
    XCircle,
    ChevronRight,
    ChevronDown,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    ChevronLeft,
  } from "@lucide/svelte";
  import type { process, processStatus } from "@/helper/model";

  let processes = $state<process[]>([]);
  let loading = $state(false);
  let expandedIds = $state<Set<string>>(new Set());

  // Filter / sort / pagination state
  let statusFilter = $state<string>("");
  let sortField = $state<string>("created_at");
  let sortDir = $state<"asc" | "desc">("desc");
  let limit = $state(25);
  let offset = $state(0);

  // Local search (client-side text filter on visible results)
  let search = $state("");

  function buildParams(): ProcessQueryParams {
    const params: ProcessQueryParams = {
      sort: sortDir === "desc" ? `-${sortField}` : sortField,
      limit,
      offset,
    };
    if (statusFilter) params.status = statusFilter;
    return params;
  }

  const fetchProcesses = async () => {
    loading = true;
    processes = await requestProcesses(buildParams());
    loading = false;
  };

  const terminateProcess = async (pid: string) => {
    try {
      await requestProcessTerminate(pid);
      addToast("Process terminated", "info");
      await fetchProcesses();
    } catch (error: any) {
      const msg =
        error?.response?.data?.error || error?.message || "Unknown error";
      addToast("Error terminating process: " + msg, "alert");
    }
  };

  const clearProcesses = async () => {
    if (!confirm("Clear all processes? This cannot be undone.")) return;
    try {
      await requestProcessDelete();
      addToast("Processes cleared", "info");
      await fetchProcesses();
    } catch (error: any) {
      const msg =
        error?.response?.data?.error || error?.message || "Unknown error";
      addToast("Error clearing processes: " + msg, "alert");
    }
  };

  const toggleExpand = (id: string) => {
    const next = new Set(expandedIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    expandedIds = next;
  };

  const toggleSort = (field: string) => {
    if (sortField === field) {
      sortDir = sortDir === "asc" ? "desc" : "asc";
    } else {
      sortField = field;
      sortDir = "desc";
    }
    offset = 0;
    fetchProcesses();
  };

  const prevPage = () => {
    offset = Math.max(0, offset - limit);
    fetchProcesses();
  };
  const nextPage = () => {
    if (processes.length === limit) {
      offset += limit;
      fetchProcesses();
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-blue-100 text-blue-700";
      case "completed":
        return "bg-green-100 text-green-700";
      case "failed":
        return "bg-red-100 text-red-700";
      case "terminated":
        return "bg-gray-200 text-gray-600";
      default:
        return "bg-gray-100 text-gray-500";
    }
  };

  const formatDate = (date: string) => {
    if (!date) return "-";
    return new Date(date).toLocaleString();
  };

  let filteredProcesses = $derived.by(() => {
    if (!search.trim()) return processes;
    const q = search.toLowerCase();
    return processes.filter(
      (p) =>
        p.id.toLowerCase().includes(q) ||
        p.status.toLowerCase().includes(q) ||
        (p.info.note ?? "").toLowerCase().includes(q) ||
        (p.info.description ?? "").toLowerCase().includes(q) ||
        (p.info.query ?? "").toLowerCase().includes(q) ||
        (p.user ?? "").toLowerCase().includes(q) ||
        (p.info.error ?? "").toLowerCase().includes(q),
    );
  });

  const sortIcon = (field: string): "neutral" | "asc" | "desc" => {
    if (sortField !== field) return "neutral";
    return sortDir;
  };

  const statuses: { label: string; value: string }[] = [
    { label: "All", value: "" },
    { label: "Running", value: "running" },
    { label: "Completed", value: "completed" },
    { label: "Failed", value: "failed" },
    { label: "Terminated", value: "terminated" },
  ];

  let initialized = $state(false);

  $effect(() => {
    // re-fetch when status filter changes (skip initial mount)
    statusFilter;
    if (initialized) {
      offset = 0;
      fetchProcesses();
    }
  });

  onMount(() => {
    fetchProcesses().then(() => {
      initialized = true;
    });
  });
</script>

<div class="flex flex-col h-full w-full overflow-hidden text-sm">
  <!-- Toolbar -->
  <div class="border-b border-gray-300 flex items-center justify-between gap-2 px-3 h-10 shrink-0 bg-gray-100">
    <div class="flex items-center gap-2">
      <span class="font-medium text-gray-700">Processes</span>
      <span class="border-l border-gray-300 h-4"></span>
      <!-- Status filter -->
      <select
        class="border-none bg-transparent text-sm hover:cursor-pointer hover:bg-white h-8 px-1"
        bind:value={statusFilter}
      >
        {#each statuses as s}
          <option value={s.value}>{s.label}</option>
        {/each}
      </select>
      <span class="border-l border-gray-300 h-4"></span>
      <!-- Search -->
      <input
        type="text"
        class="border-none bg-transparent text-sm hover:bg-white focus:bg-white h-8 px-2 w-56"
        placeholder="Search..."
        bind:value={search}
      />
    </div>
    <div class="flex items-center">
      <button
        class="px-2 h-8 hover:cursor-pointer hover:bg-red-100 text-gray-600 hover:text-red-600"
        onclick={clearProcesses}
        title="Clear all processes"
      >
        <Trash size={16} />
      </button>
      <button
        class="px-2 h-8 hover:cursor-pointer hover:bg-gray-200 text-gray-600"
        onclick={fetchProcesses}
        disabled={loading}
        title="Refresh"
      >
        <RefreshCw class={loading ? "animate-spin" : ""} size={16} />
      </button>
    </div>
  </div>

  <!-- Table -->
  <div class="flex-1 overflow-y-auto min-h-0">
    <table class="w-full text-sm">
      <thead class="sticky top-0 bg-gray-50 z-10">
        <tr class="border-b border-gray-300 text-left text-gray-500">
          <th class="px-3 py-2 w-8"></th>
          <th class="px-3 py-2 w-24">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("status")}>
              Status
              {#if sortIcon("status") === "asc"}<ArrowUp size={12} />{:else if sortIcon("status") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2">Note</th>
          <th class="px-3 py-2">Description</th>
          <th class="px-3 py-2 w-28">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("info.duration")}>
              Duration
              {#if sortIcon("info.duration") === "asc"}<ArrowUp size={12} />{:else if sortIcon("info.duration") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-20">Rows</th>
          <th class="px-3 py-2 w-24">User</th>
          <th class="px-3 py-2 w-44">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("created_at")}>
              Created
              {#if sortIcon("created_at") === "asc"}<ArrowUp size={12} />{:else if sortIcon("created_at") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-44">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("updated_at")}>
              Updated
              {#if sortIcon("updated_at") === "asc"}<ArrowUp size={12} />{:else if sortIcon("updated_at") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-20">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#if filteredProcesses.length === 0 && !loading}
          <tr>
            <td colspan="10" class="text-gray-400 text-center py-10">No processes found</td>
          </tr>
        {/if}
        {#each filteredProcesses as proc (proc.id)}
          {@const isExpanded = expandedIds.has(proc.id)}
          {@const hasDetail = proc.info.query || proc.info.error}
          <tr
            class={[
              "border-b border-gray-200 transition-colors duration-100",
              hasDetail ? "hover:bg-gray-50 cursor-pointer" : "",
              isExpanded ? "bg-gray-50" : "",
            ]}
            onclick={() => { if (hasDetail) toggleExpand(proc.id); }}
          >
            <td class="px-3 py-2.5 text-gray-400">
              {#if hasDetail}
                {#if isExpanded}
                  <ChevronDown size={14} />
                {:else}
                  <ChevronRight size={14} />
                {/if}
              {/if}
            </td>
            <td class="px-3 py-2.5">
              <span class={["px-2 py-0.5 text-xs font-medium rounded", statusColor(proc.status)]}>
                {proc.status}
              </span>
            </td>
            <td class="px-3 py-2.5 truncate max-w-56" title={proc.info.note}>
              {proc.info.note || "-"}
            </td>
            <td class="px-3 py-2.5 truncate max-w-64" title={proc.info.description}>
              {proc.info.description || "-"}
            </td>
            <td class="px-3 py-2.5 tabular-nums">
              {proc.info.duration || "-"}
            </td>
            <td class="px-3 py-2.5 tabular-nums">
              {proc.info.rows_affected ?? "-"}
            </td>
            <td class="px-3 py-2.5 truncate max-w-24" title={proc.user}>
              {proc.user || "-"}
            </td>
            <td class="px-3 py-2.5 whitespace-nowrap tabular-nums">
              {formatDate(proc.created_at)}
            </td>
            <td class="px-3 py-2.5 whitespace-nowrap tabular-nums">
              {formatDate(proc.updated_at)}
            </td>
            <td class="px-3 py-2.5">
              {#if proc.status === "running"}
                <button
                  class="text-red-500 hover:text-white hover:bg-red-500 p-1 rounded hover:cursor-pointer"
                  onclick={(e: MouseEvent) => { e.stopPropagation(); terminateProcess(proc.id); }}
                  title="Terminate"
                >
                  <XCircle size={16} />
                </button>
              {/if}
            </td>
          </tr>
          {#if isExpanded}
            <tr class="border-b border-gray-200">
              <td colspan="10" class="p-0">
                <div class="border-l-2 border-gray-300 ml-5 my-2">
                  {#if proc.info.error}
                    <div class="px-4 py-2 bg-red-50 text-red-700 text-sm">
                      <span class="font-medium">Error:</span> {proc.info.error}
                    </div>
                  {/if}
                  {#if proc.info.query}
                    <div class="px-4 py-2">
                      <span class="text-gray-400 text-xs uppercase tracking-wide">Query</span>
                      <pre class="mt-1 text-sm text-gray-700 whitespace-pre-wrap break-all bg-white border border-gray-200 rounded px-3 py-2 max-h-72 overflow-y-auto">{proc.info.query}</pre>
                    </div>
                  {/if}
                </div>
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  <div class="border-t border-gray-300 flex items-center justify-between px-3 h-10 shrink-0 bg-gray-100 text-sm text-gray-500">
    <div class="flex items-center gap-2">
      <span>Showing {offset + 1}-{offset + filteredProcesses.length}</span>
      <span class="border-l border-gray-300 h-4"></span>
      <label class="flex items-center gap-1">
        Per page
        <select
          class="border-none bg-transparent hover:cursor-pointer hover:bg-white h-7 text-sm px-1"
          bind:value={limit}
          onchange={() => { offset = 0; fetchProcesses(); }}
        >
          <option value={10}>10</option>
          <option value={25}>25</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </label>
    </div>
    <div class="flex items-center gap-1">
      <button
        class="px-2 h-7 hover:bg-gray-200 hover:cursor-pointer disabled:opacity-30 disabled:cursor-default rounded"
        disabled={offset === 0}
        onclick={prevPage}
        title="Previous page"
      >
        <ChevronLeft size={14} />
      </button>
      <button
        class="px-2 h-7 hover:bg-gray-200 hover:cursor-pointer disabled:opacity-30 disabled:cursor-default rounded"
        disabled={processes.length < limit}
        onclick={nextPage}
        title="Next page"
      >
        <ChevronRight size={14} />
      </button>
    </div>
  </div>
</div>
