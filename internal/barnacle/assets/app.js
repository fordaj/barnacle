/* barnacle page shell: theme toggle + client-side search.
 *
 * Theme: System / Light / Dark, default System, persisted in
 * localStorage under "barnacle-theme". The head of each page carries a
 * tiny inline script that applies the stored preference before first
 * paint (no flash of the wrong theme); this file only wires up the
 * toggle button. The visible icon is selected in pure CSS off the
 * <html data-theme="…"> attribute, so the button works with styles
 * alone once data-theme is set.
 *
 * Search: barnacle is a static site, so searching happens in the
 * browser against window.BARNACLE_SEARCH_INDEX, generated at build
 * time by barnacle (see search.go). The index is loaded with a plain
 * <script src>, which works on file:// URLs where fetch() of a JSON
 * file does not. Matching is case-insensitive substring AND across
 * whitespace-separated terms, ranked title > heading > body; results
 * link to page.html#section-anchor.
 */
(function () {
  "use strict";

  /* ---------- Sidebar toggle ---------- */
  var SIDEBAR_KEY = "barnacle-sidebar";
  var sidebarToggle = document.getElementById("sidebar-toggle");

  function applySidebar(collapsed, persist) {
    document.documentElement.setAttribute(
      "data-sidebar",
      collapsed ? "collapsed" : "open",
    );
    if (sidebarToggle) {
      sidebarToggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
      sidebarToggle.setAttribute(
        "title",
        collapsed ? "Show sidebar" : "Hide sidebar",
      );
    }
    if (persist) {
      try {
        localStorage.setItem(SIDEBAR_KEY, collapsed ? "collapsed" : "open");
      } catch (e) {
        /* storage unavailable — the toggle still applies for this page view */
      }
    }
  }

  if (sidebarToggle) {
    sidebarToggle.addEventListener("click", function () {
      var collapsed =
        document.documentElement.getAttribute("data-sidebar") !== "collapsed";
      applySidebar(collapsed, true);
    });
    applySidebar(
      document.documentElement.getAttribute("data-sidebar") === "collapsed",
      false,
    );
  }

  /* ---------- Theme toggle ---------- */
  var THEME_KEY = "barnacle-theme";
  var THEME_ORDER = ["system", "light", "dark"];
  var toggle = document.getElementById("theme-toggle");

  function currentTheme() {
    var t = document.documentElement.getAttribute("data-theme");
    return THEME_ORDER.indexOf(t) === -1 ? "system" : t;
  }

  function describeTheme(theme) {
    return (
      "Theme: " +
      theme.charAt(0).toUpperCase() +
      theme.slice(1) +
      " — activate to change"
    );
  }

  function applyTheme(theme, persist) {
    document.documentElement.setAttribute("data-theme", theme);
    if (toggle) {
      toggle.setAttribute("aria-label", describeTheme(theme));
      toggle.setAttribute("title", describeTheme(theme));
    }
    if (persist) {
      try {
        localStorage.setItem(THEME_KEY, theme);
      } catch (e) {
        /* private mode etc. — the toggle still applies for this page view */
      }
    }
  }

  if (toggle) {
    applyTheme(currentTheme(), false);
    toggle.addEventListener("click", function () {
      var next =
        THEME_ORDER[
          (THEME_ORDER.indexOf(currentTheme()) + 1) % THEME_ORDER.length
        ];
      applyTheme(next, true);
    });
  }

  /* ---------- Search ---------- */
  var container = document.getElementById("search");
  var input = document.getElementById("barnacle-search");
  var results = document.getElementById("search-results");
  if (!container || !input || !results) {
    return;
  }

  var prefix = container.getAttribute("data-prefix") || "";
  var index = Array.isArray(window.BARNACLE_SEARCH_INDEX)
    ? window.BARNACLE_SEARCH_INDEX
    : [];

  // Lowercased searchable blob per page, computed once.
  for (var i = 0; i < index.length; i++) {
    var page = index[i];
    var blob = (page.t || "").toLowerCase();
    for (var j = 0; j < page.s.length; j++) {
      blob +=
        "\n" +
        (page.s[j].t || "").toLowerCase() +
        "\n" +
        (page.s[j].x || "").toLowerCase();
    }
    page._blob = blob;
  }

  var MAX_PAGES = 8;
  var MAX_SECTIONS = 3;

  function occurrences(haystack, needle) {
    if (!needle) {
      return 0;
    }
    var count = 0;
    var pos = 0;
    while ((pos = haystack.indexOf(needle, pos)) !== -1) {
      count++;
      pos += needle.length;
    }
    return count;
  }

  function rank(query) {
    var terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    if (terms.length === 0) {
      return [];
    }
    var out = [];
    for (var i = 0; i < index.length; i++) {
      var page = index[i];
      var matchesAll = true;
      for (var t = 0; t < terms.length; t++) {
        if (page._blob.indexOf(terms[t]) === -1) {
          matchesAll = false;
          break;
        }
      }
      if (!matchesAll) {
        continue;
      }
      var pageScore = 0;
      for (var u = 0; u < terms.length; u++) {
        if ((page.t || "").toLowerCase().indexOf(terms[u]) !== -1) {
          pageScore += 20;
        }
      }
      var hits = [];
      for (var s = 0; s < page.s.length; s++) {
        var section = page.s[s];
        var lt = (section.t || "").toLowerCase();
        var lx = (section.x || "").toLowerCase();
        var score = 0;
        for (var k = 0; k < terms.length; k++) {
          if (lt.indexOf(terms[k]) !== -1) {
            score += 10;
          }
          score += occurrences(lx, terms[k]);
        }
        if (score > 0) {
          hits.push({ section: section, score: score });
        }
      }
      hits.sort(function (a, b) {
        return b.score - a.score;
      });
      out.push({
        page: page,
        hits: hits,
        total: pageScore + (hits.length > 0 ? hits[0].score : 0),
      });
    }
    out.sort(function (a, b) {
      return b.total - a.total;
    });
    return out.slice(0, MAX_PAGES);
  }

  function escapeRegExp(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  function appendMarked(parent, text, terms) {
    var re = new RegExp("(" + terms.map(escapeRegExp).join("|") + ")", "ig");
    var last = 0;
    var m;
    while ((m = re.exec(text)) !== null) {
      if (m.index > last) {
        parent.appendChild(document.createTextNode(text.slice(last, m.index)));
      }
      var mark = document.createElement("mark");
      mark.className = "search-mark";
      mark.textContent = m[0];
      parent.appendChild(mark);
      last = m.index + m[0].length;
    }
    if (last < text.length) {
      parent.appendChild(document.createTextNode(text.slice(last)));
    }
  }

  // appendRichText appends text to parent, rendering single-backtick
  // spans as <code> elements and wrapping occurrences of terms in
  // <mark>. Everything passes through textContent/createTextNode, so
  // markup-bearing input cannot inject HTML. In text contexts with no
  // query (titles), terms is empty and backticks alone are honored.
  function appendRichText(parent, text, terms) {
    var chunks = String(text).split(/`([^`]+)`/g);
    for (var i = 0; i < chunks.length; i++) {
      if (i % 2 === 1) {
        var code = document.createElement("code");
        code.textContent = chunks[i];
        parent.appendChild(code);
      } else if (chunks[i]) {
        if (terms && terms.length) {
          appendMarked(parent, chunks[i], terms);
        } else {
          parent.appendChild(document.createTextNode(chunks[i]));
        }
      }
    }
  }

  function snippet(text, terms) {
    var lower = text.toLowerCase();
    var at = -1;
    for (var i = 0; i < terms.length; i++) {
      var idx = lower.indexOf(terms[i]);
      if (idx !== -1 && (at === -1 || idx < at)) {
        at = idx;
      }
    }
    if (at === -1) {
      at = 0;
    }
    var start = Math.max(0, at - 80);
    var end = Math.min(text.length, at + 160);
    return (
      (start > 0 ? "…" : "") +
      text.slice(start, end) +
      (end < text.length ? "…" : "")
    );
  }

  function setExpanded(expanded) {
    input.setAttribute("aria-expanded", expanded ? "true" : "false");
  }

  function clearResults() {
    while (results.firstChild) {
      results.removeChild(results.firstChild);
    }
    results.hidden = true;
    setExpanded(false);
  }

  function render(query) {
    var terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    var matches = rank(query);
    clearResults();
    if (terms.length === 0) {
      return;
    }
    if (matches.length === 0) {
      var empty = document.createElement("div");
      empty.className = "search-empty";
      empty.textContent = "No results for “" + query + "”.";
      results.appendChild(empty);
      results.hidden = false;
      setExpanded(true);
      return;
    }
    for (var i = 0; i < matches.length; i++) {
      var match = matches[i];
      var group = document.createElement("div");
      group.className = "search-group";

      var pageLink = document.createElement("a");
      pageLink.className = "search-page";
      pageLink.setAttribute("role", "option");
      pageLink.href = prefix + match.page.u;
      appendRichText(pageLink, match.page.t, []);
      group.appendChild(pageLink);

      var limit = Math.min(match.hits.length, MAX_SECTIONS);
      for (var j = 0; j < limit; j++) {
        var hit = match.hits[j];
        var a = document.createElement("a");
        a.className = "search-hit";
        a.setAttribute("role", "option");
        a.href =
          prefix + match.page.u + (hit.section.a ? "#" + hit.section.a : "");

        var title = document.createElement("span");
        title.className = "search-hit-title";
        appendRichText(title, hit.section.a ? hit.section.t : match.page.t, []);
        a.appendChild(title);

        var snip = document.createElement("span");
        snip.className = "search-snippet";
        appendRichText(
          snip,
          snippet(hit.section.x || hit.section.t, terms),
          terms,
        );
        a.appendChild(snip);

        group.appendChild(a);
      }
      results.appendChild(group);
    }
    results.hidden = false;
    setExpanded(true);
  }

  var activeIndex = -1;

  function resultLinks() {
    return Array.prototype.slice.call(results.querySelectorAll("a"));
  }

  function setActive(index_) {
    var list = resultLinks();
    if (list.length === 0) {
      return;
    }
    var next = ((index_ % list.length) + list.length) % list.length;
    activeIndex = next;
    for (var i = 0; i < list.length; i++) {
      list[i].setAttribute("data-active", i === next ? "true" : "false");
    }
    list[next].scrollIntoView({ block: "nearest" });
  }

  input.addEventListener("input", function () {
    activeIndex = -1;
    render(input.value);
  });

  input.addEventListener("focus", function () {
    if (input.value.trim()) {
      render(input.value);
    }
  });

  input.addEventListener("keydown", function (event) {
    if (results.hidden) {
      if (event.key === "Escape") {
        input.blur();
      }
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive(activeIndex + 1);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive(activeIndex - 1);
    } else if (event.key === "Enter" && activeIndex >= 0) {
      var list = resultLinks();
      if (list[activeIndex]) {
        event.preventDefault();
        list[activeIndex].click();
      }
    } else if (event.key === "Escape") {
      event.preventDefault();
      input.value = "";
      clearResults();
      input.blur();
    }
  });

  // Keep focus inside the dropdown row (mousedown) without stealing
  // the click that follows.
  results.addEventListener("mousedown", function (event) {
    event.preventDefault();
  });

  container.addEventListener("focusout", function (event) {
    if (!event.relatedTarget || !container.contains(event.relatedTarget)) {
      clearResults();
    }
  });

  document.addEventListener("keydown", function (event) {
    var meta = event.metaKey || event.ctrlKey;
    var target = event.target;
    var editable =
      target &&
      (/^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName) ||
        target.isContentEditable);
    if (meta && (event.key === "k" || event.key === "K")) {
      event.preventDefault();
      input.focus();
      input.select();
      return;
    }
    if (event.key === "/" && !editable) {
      event.preventDefault();
      input.focus();
      input.select();
    }
  });
})();
