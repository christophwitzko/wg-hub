"use client";

import { useEffect, useMemo, useState } from "react";
import { HighlighterCore } from "shiki";
import { getHighlighterCore } from "shiki/core";

import getWasm from "shiki/wasm";
import ghDark from "shiki/themes/github-dark.mjs";
import ghLight from "shiki/themes/github-light.mjs";
import yaml from "shiki/langs/yaml.mjs";

import { cn } from "@/lib/utils";
import { useTheme } from "next-themes";

export function Code({
  lang = "yaml",
  value,
  className,
}: {
  lang: "yaml";
  value: string;
  className?: string;
}) {
  const { resolvedTheme } = useTheme();
  const [highlighter, setHighlighter] = useState<HighlighterCore | null>(null);
  useEffect(() => {
    getHighlighterCore({
      themes: [ghDark, ghLight],
      langs: [yaml],
      loadWasm: getWasm,
    }).then((highlighter) => {
      setHighlighter(highlighter);
    });
  }, []);
  const html = useMemo(
    () =>
      highlighter
        ? highlighter.codeToHtml(value, {
            lang: lang,
            theme: resolvedTheme === "dark" ? "github-dark" : "github-light",
          })
        : "",
    [resolvedTheme, highlighter, lang, value],
  );
  return (
    <div
      className={cn("[&>pre]:p-2 [&>pre]:h-full border", className)}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
