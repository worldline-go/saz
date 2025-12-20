export const tableHTML = (output: { rows?: string[][]; columns: string[] } | null): string => {
  if (!output) {
    return '<table><thead></thead><tbody></tbody></table>';
  }

  if (!output.rows) {
    output.rows = [];
  }

  const headerRow = '<th class="row-number-header">#</th>' + output.columns
    .map((col, idx) => `<th onclick="sortTable(${idx}, event)" class="sortable">${col} <span class="sort-icon"></span><span class="sort-order"></span></th>`)
    .join('');

  const bodyRows = output.rows
    .map((row, rowIndex) => {
      const rowData = row
        .map(cell => `<td title="${cell}" onclick="handleCellClick(event, this)" style="cursor: pointer;">${cell}</td>`)
        .join('');
      return `<tr><td class="row-number">${rowIndex + 1}</td>${rowData}</tr>`;
    })
    .join('');

  return `
    <style>
      .table-container {
        padding: 10px;
        background: #ffffff;
      }
      
      .search-container {
        margin-bottom: 10px;
        display: flex;
        gap: 8px;
      }
      
      .search-input {
        flex: 1;
        padding: 6px 10px;
        border: 1px solid #ccc;
        border-radius: 4px;
        font-size: 14px;
      }
      
      .search-input:focus {
        outline: none;
        border-color: #666;
      }
      
      .search-clear {
        padding: 6px 12px;
        background: #f0f0f0;
        border: 1px solid #ccc;
        border-radius: 4px;
        cursor: pointer;
        font-size: 14px;
      }
      
      .search-clear:hover {
        background: #e0e0e0;
      }
      
      .table-wrapper {
        border: 1px solid #ddd;
        overflow-x: auto;
        overflow-y: auto;
      }
      
      .data-table {
        width: max-content;
        min-width: 100%;
        border-collapse: collapse;
        font-size: 13px;
      }
      
      .data-table thead {
        background: #f5f5f5;
        position: sticky;
        top: 0;
        z-index: 10;
      }
      
      .data-table th {
        padding: 10px 12px;
        text-align: left;
        font-weight: 600;
        border-bottom: 2px solid #ddd;
        border-right: 1px solid #ddd;
        cursor: pointer;
        user-select: none;
        white-space: nowrap;
        min-width: 100px;
      }
      
      .data-table th.row-number-header {
        cursor: default;
        width: 20px;
        min-width: 10px;
        text-align: center;
        background: #e8e8e8;
        position: sticky;
        left: 0;
        z-index: 11;
      }
      
      .data-table td.row-number {
        text-align: center;
        font-weight: 500;
        color: #666;
        background: #fafafa;
        width: 20px;
        min-width: 10px;
        position: sticky;
        left: 0;
        z-index: 5;
      }
      
      .data-table th.sortable:hover {
        background: #e8e8e8;
      }
      
      .data-table th .sort-icon {
        margin-left: 4px;
        font-size: 11px;
        display: inline-block;
        min-width: 12px;
      }
      
      .data-table th .sort-icon::before {
        content: '\\2195';
        opacity: 0.4;
      }
      
      .data-table th.sorted-asc .sort-icon::before {
        content: '\\25B2';
        opacity: 1;
      }
      
      .data-table th.sorted-desc .sort-icon::before {
        content: '\\25BC';
        opacity: 1;
      }
      
      .data-table th .sort-order {
        margin-left: 2px;
        font-size: 9px;
        color: #666;
        font-weight: bold;
      }
      
      .data-table tbody tr {
        border-bottom: 1px solid #eee;
      }
      
      .data-table tbody tr:hover {
        background: #fff085;
      }
      
      .data-table tbody tr:nth-child(even) {
        background: #fafafa;
      }
      
      .data-table tbody tr:nth-child(even):hover {
        background: #fff085;
      }
      
      .data-table td {
        padding: 8px 12px;
        min-width: 100px;
        max-width: 400px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        border-right: 1px solid #eee;
      }
      
      .data-table td:hover {
        overflow: visible;
        white-space: normal;
        word-wrap: break-word;
        position: relative;
        z-index: 20;
        background: #fffacd;
        box-shadow: 0 2px 8px rgba(0,0,0,0.15);
      }
      
      .data-table tbody tr:hover td.row-number {
        background: #fff085;
      }
      
      .data-table tbody tr:nth-child(even):hover td.row-number {
        background: #fff085;
      }
      
      .table-info {
        margin-bottom: 10px;
        padding: 8px;
        background: #f9f9f9;
        border: 1px solid #ddd;
        font-size: 12px;
        display: flex;
        justify-content: space-between;
      }
    </style>
    
    <div class="table-container">
      <div class="table-info">
        <span>Total: <strong id="totalRows">${output.rows.length}</strong></span>
        <span style="font-size: 11px; color: #666;">Click columns to sort (supports multiple columns)</span>
      </div>
      <div class="search-container">
        <input 
          type="text" 
          id="searchInput" 
          class="search-input" 
          placeholder="Search..." 
          oninput="filterTable()"
        />
        <button class="search-clear" onclick="clearSearch()">Clear</button>
        <button class="search-clear" onclick="resetSort()">Reset Sort</button>
      </div>
      
      <div class="table-wrapper">
        <table class="data-table" id="dataTable">
          <thead>
            <tr>${headerRow}</tr>
          </thead>
          <tbody id="tableBody">
            ${bodyRows}
          </tbody>
        </table>
      </div>
    </div>
    
    <script>
      let sortColumns = [];  // Array of {column, direction}
      let originalOrder = [];
      
      // Store original order on load
      window.addEventListener('DOMContentLoaded', function() {
        const tbody = document.getElementById('tableBody');
        if (tbody) {
          originalOrder = Array.from(tbody.getElementsByTagName('tr'));
        }
      });
      
      function sortTable(columnIndex, event) {
        const table = document.getElementById('dataTable');
        const tbody = document.getElementById('tableBody');
        const rows = Array.from(tbody.getElementsByTagName('tr'));
        const headers = table.getElementsByTagName('th');
        
        // Store original order if not already stored
        if (originalOrder.length === 0) {
          originalOrder = Array.from(rows);
        }
        
        // Find if this column is already in sort list
        const existingIndex = sortColumns.findIndex(s => s.column === columnIndex);
        
        // Toggle sort direction for this column
        let direction = 'asc';
        if (existingIndex >= 0) {
          const current = sortColumns[existingIndex].direction;
          if (current === 'asc') {
            direction = 'desc';
          } else {
            // Remove this column from sort
            sortColumns.splice(existingIndex, 1);
            updateSortUI(headers);
            if (sortColumns.length === 0) {
              originalOrder.forEach(row => tbody.appendChild(row));
            } else {
              performSort(rows, tbody);
            }
            return;
          }
          sortColumns[existingIndex].direction = direction;
        } else {
          sortColumns.push({ column: columnIndex, direction: direction });
        }
        
        // Update UI
        updateSortUI(headers);
        
        // Perform sort
        performSort(rows, tbody);
      }
      
      function updateSortUI(headers) {
        // Clear all sort classes and order numbers
        for (let i = 0; i < headers.length; i++) {
          headers[i].classList.remove('sorted-asc', 'sorted-desc');
          const orderSpan = headers[i].querySelector('.sort-order');
          if (orderSpan) orderSpan.textContent = '';
        }
        
        // Add classes and order numbers for sorted columns
        sortColumns.forEach((sort, index) => {
          // +1 offset to skip the row number header (first header)
          headers[sort.column + 1].classList.add('sorted-' + sort.direction);
          const orderSpan = headers[sort.column + 1].querySelector('.sort-order');
          if (orderSpan && sortColumns.length > 1) {
            orderSpan.textContent = (index + 1);
          }
        });
      }
      
      function performSort(rows, tbody) {
        rows.sort((a, b) => {
          const aCells = a.getElementsByTagName('td');
          const bCells = b.getElementsByTagName('td');
          
          // Compare by each sort column in order
          for (let i = 0; i < sortColumns.length; i++) {
            const sort = sortColumns[i];
            // +1 offset to skip the row number column (first column)
            const aValue = aCells[sort.column + 1].textContent;
            const bValue = bCells[sort.column + 1].textContent;
            
            // Try to parse as numbers
            const aNum = parseFloat(aValue);
            const bNum = parseFloat(bValue);
            
            let comparison = 0;
            if (!isNaN(aNum) && !isNaN(bNum)) {
              comparison = aNum - bNum;
            } else {
              comparison = aValue.localeCompare(bValue);
            }
            
            if (comparison !== 0) {
              return sort.direction === 'asc' ? comparison : -comparison;
            }
            // If equal, continue to next sort column
          }
          return 0;
        });
        
        // Reorder DOM
        rows.forEach(row => tbody.appendChild(row));
      }
      
      function filterTable() {
        const input = document.getElementById('searchInput');
        const filter = input.value.toLowerCase();
        const tbody = document.getElementById('tableBody');
        const rows = tbody.getElementsByTagName('tr');
        let visibleCount = 0;
        
        for (let i = 0; i < rows.length; i++) {
          const row = rows[i];
          const cells = row.getElementsByTagName('td');
          let found = false;
          
          for (let j = 0; j < cells.length; j++) {
            const cellText = cells[j].textContent.toLowerCase();
            if (cellText.includes(filter)) {
              found = true;
              break;
            }
          }
          
          if (found) {
            row.style.display = '';
            visibleCount++;
          } else {
            row.style.display = 'none';
          }
        }
        
        document.getElementById('visibleRows').textContent = visibleCount;
      }
      
      function clearSearch() {
        document.getElementById('searchInput').value = '';
        filterTable();
      }
      
      function resetSort() {
        const table = document.getElementById('dataTable');
        const tbody = document.getElementById('tableBody');
        const headers = table.getElementsByTagName('th');
        
        sortColumns = [];
        updateSortUI(headers);
        originalOrder.forEach(row => tbody.appendChild(row));
      }
      
      function handleCellClick(event, cell) {
        if (event.ctrlKey || event.metaKey) {
          const value = cell.textContent;
          navigator.clipboard.writeText(value);
          // Visual feedback
          const originalBg = cell.style.backgroundColor;
          cell.style.backgroundColor = '#90EE90';
          setTimeout(() => {
            cell.style.backgroundColor = originalBg;
          }, 200);
        }
      }
    </script>
  `;
};
