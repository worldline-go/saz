export const outputToData = (output: { rows?: string[][]; columns: string[] } | null): any[] => {
  if (!output || !output.columns) {
    return [];
  }

  if (!output.rows) {
    output.rows = [];
  }

  return output.rows.map(row => {
    const record: any = {};
    output.columns.forEach((col, index) => {
      record[col] = row[index];
    });

    return record;
  });
};

export const exportToJSON = (data: any[], filename: string): void => {
  const jsonContent = JSON.stringify(data, null, 2);
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  link.setAttribute('href', url);
  link.setAttribute('download', filename);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
};

export const exportToCSV = (output: { rows?: string[][]; columns: string[] } | null, filename: string): void => {
  if (!output || !output.columns) {
    return;
  }

  if (!output.rows) {
    output.rows = [];
  }

  const csvRows: string[] = [];
  csvRows.push(output.columns.map(col => `"${col.replace(/"/g, '""')}"`).join(','));

  output.rows.forEach(row => {
    const csvRow = row.map(cell => `"${(cell ?? '').replace(/"/g, '""')}"`).join(',');
    csvRows.push(csvRow);
  });

  const csvContent = csvRows.join('\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  link.setAttribute('href', url);
  link.setAttribute('download', filename);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
};
