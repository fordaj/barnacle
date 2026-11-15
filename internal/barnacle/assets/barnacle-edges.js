/*
 * barnacle edge links.
 *
 * Mermaid can click nodes but not edges. Barnacle lets authors attach
 * links to edges with a comment directive in the diagram source:
 *
 *   %% barnacle: link e1 href "../protocol.md#send-cmd" "optional tooltip" [_blank]
 *
 *   key (bare)    match the edge's data-id, e.g. the explicit e1 in
 *                 `A e1@--> B`, or the auto-assigned id of any other
 *                 edge (`edge0`, `edge1`, … in state diagrams; `L_A_B_0`
 *                 in unnamed flowcharts).
 *   key (quoted)  match the edge's label text instead, e.g.
 *                 %% barnacle: link "send(cmd)" href "dispatch.md"
 *   href          a link target; .md/.markdown/.mmd targets are rewritten
 *                 to the generated .html pages by barnacle at build time.
 *
 * Barnacle parses these directives at build time (the directive lines are
 * mermaid comments and render nothing), rewrites the targets, and emits a
 * JSON manifest right after each diagram:
 *
 *   <script type="application/json" class="barnacle-edge-links">{"edges":[…]}</script>
 *
 * This script runs after mermaid renders (MutationObserver, so script order
 * does not matter), finds each diagram's manifest, and wraps the matching
 * edge strokes — including an invisible, wide copy of the path for a real
 * hit area — in an SVG <a> element. Keyboard navigation (tab, Enter/Space)
 * and a visible focus outline are built in.
 */
(function () {
  "use strict";
  if (typeof document === "undefined") {
    return; // SSR guard; there is nothing to wire up.
  }

  var SVG_NS = "http://www.w3.org/2000/svg";
  var XLINK_NS = "http://www.w3.org/1999/xlink";

  function norm(text) {
    return (text || "").replace(/\s+/g, " ").trim();
  }

  function cssEscape(value) {
    return String(value).replace(/(["\\])/g, "\\$1");
  }

  function resolveEdge(svg, entry) {
    if (entry.by === "label") {
      var labels = svg.querySelectorAll("g.label[data-id]");
      for (var i = 0; i < labels.length; i++) {
        if (norm(labels[i].textContent) === norm(entry.key)) {
          var dataID = labels[i].getAttribute("data-id");
          return svg.querySelector('path[data-id="' + cssEscape(dataID) + '"]');
        }
      }
      return null;
    }
    return svg.querySelector('path[data-id="' + cssEscape(entry.key) + '"]');
  }

  function navigate(entry) {
    if (entry.blank) {
      window.open(entry.href, "_blank", "noopener");
    } else {
      window.location.assign(entry.href);
    }
  }

  function wireEdge(path, entry, duplicate) {
    if (duplicate) {
      console.warn(
        "barnacle: multiple edge links for key " +
          entry.key +
          "; first one wins",
      );
      return;
    }
    var parent = path.parentNode;
    if (!parent) {
      return;
    }
    var a = document.createElementNS(SVG_NS, "a");
    a.setAttributeNS(XLINK_NS, "href", entry.href);
    if (entry.blank) {
      a.setAttribute("target", "_blank");
      a.setAttribute("rel", "noopener");
    }
    a.setAttribute("class", "barnacle-edge-link");
    a.setAttribute("tabindex", "0");
    a.setAttribute("role", "link");
    a.setAttribute("aria-label", entry.tooltip || entry.href);
    // Keyboard fallback: SVG anchors are not guaranteed to activate on
    // Enter/Space everywhere, so drive navigation from the key handler.
    a.addEventListener("keydown", function (event) {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        navigate(entry);
      }
    });
    parent.replaceChild(a, path);
    a.appendChild(path);
    // Invisible, generously wide copy of the stroke so the edge is
    // actually clickable, not just its 2px of ink.
    var hit = path.cloneNode(false);
    hit.removeAttribute("id");
    hit.removeAttribute("class");
    hit.removeAttribute("data-id");
    hit.removeAttribute("data-points");
    hit.removeAttribute("data-et");
    hit.removeAttribute("data-edge");
    hit.removeAttribute("marker-start");
    hit.removeAttribute("marker-end");
    hit.setAttribute("class", "barnacle-edge-hit");
    hit.setAttribute("fill", "none");
    hit.setAttribute("stroke", "#000");
    hit.setAttribute("stroke-opacity", "0");
    hit.setAttribute("stroke-width", "16");
    hit.setAttribute("pointer-events", "stroke");
    a.appendChild(hit);
  }

  function processDiagram(pre) {
    if (pre.getAttribute("data-barnacle-edge-links") === "done") {
      return;
    }
    var manifestScript = null;
    for (
      var node = pre.nextElementSibling;
      node;
      node = node.nextElementSibling
    ) {
      if (node.classList && node.classList.contains("barnacle-edge-links")) {
        manifestScript = node;
        break;
      }
      // Never skip past content that is not a barnacle manifest.
      if (node.tagName && node.tagName.toLowerCase() !== "script") {
        break;
      }
    }
    if (!manifestScript) {
      // The manifest is static; if none is next to this diagram there
      // never will be. Mark done so scans stop revisiting it.
      pre.setAttribute("data-barnacle-edge-links", "done");
      return;
    }
    var svg = pre.querySelector("svg");
    if (!svg) {
      return;
    }
    if (!svg.querySelector("path[data-id]")) {
      // Mermaid marks the block processed before the SVG is populated;
      // wait for actual edge paths instead of wiring an empty skeleton.
      return;
    }
    var manifest;
    try {
      manifest = JSON.parse(manifestScript.textContent);
    } catch (err) {
      console.warn("barnacle: unreadable edge links manifest", err);
      return;
    }
    var edges = manifest && Array.isArray(manifest.edges) ? manifest.edges : [];
    var seen = Object.create(null);
    for (var i = 0; i < edges.length; i++) {
      var entry = edges[i] || {};
      if (!entry.key || !entry.href) {
        continue;
      }
      var duplicate = Object.prototype.hasOwnProperty.call(seen, entry.key);
      seen[entry.key] = true;
      var path = resolveEdge(svg, entry);
      if (!path) {
        console.warn("barnacle: no edge matches key " + entry.key);
        continue;
      }
      wireEdge(path, entry, duplicate);
    }
    pre.setAttribute("data-barnacle-edge-links", "done");
  }

  function scan() {
    var pres = document.querySelectorAll("pre.mermaid[data-processed='true']");
    for (var i = 0; i < pres.length; i++) {
      processDiagram(pres[i]);
    }
  }

  function start() {
    var observer = new MutationObserver(function (mutations) {
      var interesting = false;
      for (var i = 0; i < mutations.length; i++) {
        if (mutations[i].addedNodes.length > 0) {
          interesting = true;
          break;
        }
      }
      if (interesting) {
        scan();
      }
    });
    observer.observe(document.body, { subtree: true, childList: true });
    // Also process any diagrams that finished rendering before we ran.
    scan();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start);
  } else {
    start();
  }
})();
