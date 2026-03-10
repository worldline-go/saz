const DANGEROUS_KEYWORDS = ["DELETE", "DROP", "TRUNCATE", "ALTER", "UPDATE"];

/**
 * Scans SQL text for dangerous keywords using whole-word matching.
 * Returns the list of matched keywords (uppercased), or an empty array if none found.
 */
export function getDangerousKeywords(sql: string): string[] {
  const matches: string[] = [];
  for (const keyword of DANGEROUS_KEYWORDS) {
    const re = new RegExp(`\\b${keyword}\\b`, "i");
    if (re.test(sql)) {
      matches.push(keyword);
    }
  }
  return matches;
}

/**
 * If the SQL contains dangerous keywords, prompts the user for confirmation.
 * Returns true if safe to proceed (no dangerous keywords or user confirmed).
 */
export function confirmDangerousQuery(sql: string): boolean {
  const keywords = getDangerousKeywords(sql);
  if (keywords.length === 0) return true;

  return confirm(
    `This query contains dangerous operations: ${keywords.join(", ")}\n\nAre you sure you want to run it?`,
  );
}
