/* ==========================================================================
   TOURCHEF — static prototype interactions
   No backend: every action gives visible feedback (toast / state change).
   ========================================================================== */
(function () {
  'use strict';

  function qs(s, r) { return (r || document).querySelector(s); }
  function qsa(s, r) { return Array.prototype.slice.call((r || document).querySelectorAll(s)); }

  /* ----- Toast ----- */
  var host = document.createElement('div');
  host.className = 'toast-host';
  document.body.appendChild(host);
  function toast(msg) {
    var t = document.createElement('div');
    t.className = 'toast';
    t.textContent = msg;
    host.appendChild(t);
    setTimeout(function () { t.style.opacity = '0'; t.style.transition = 'opacity .3s'; setTimeout(function () { t.remove(); }, 300); }, 2600);
  }
  window.tcToast = toast;

  /* ----- Header mobile menu ----- */
  var toggle = qs('.menu-toggle');
  var menu = qs('.mobile-menu');
  if (toggle && menu) {
    toggle.addEventListener('click', function () {
      toggle.classList.toggle('open');
      menu.classList.toggle('open');
    });
  }

  /* ----- Generic toggle groups (chips / tabs / seg) ----- */
  qsa('[data-toggle-group]').forEach(function (group) {
    group.addEventListener('click', function (e) {
      var btn = e.target.closest('button');
      if (!btn || !group.contains(btn)) return;
      qsa('button', group).forEach(function (b) { b.classList.remove('on'); });
      btn.classList.add('on');
      var target = group.getAttribute('data-toggle-group');
      var panes = qsa('[data-pane]');
      panes.forEach(function (p) { p.style.display = p.getAttribute('data-pane') === btn.dataset.val ? '' : 'none'; });
      if (target) toast(target + ': ' + (btn.textContent || '').trim());
    });
  });

  /* ----- Like / save / follow toggles ----- */
  qsa('[data-count]').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var out = qs('[data-count-out]', btn) || btn.querySelector('span');
      var n = parseInt(btn.getAttribute('data-count'), 10) || 0;
      var on = btn.classList.toggle('on');
      if (out) out.textContent = String(n + (on ? 1 : 0));
      var label = btn.getAttribute('data-label') || 'Saqlangan';
      toast(on ? label + " qo'shildi" : label + " o'chirildi");
    });
  });

  /* ----- Modal ----- */
  qsa('[data-modal-open]').forEach(function (b) {
    b.addEventListener('click', function (e) {
      e.preventDefault();
      var m = qs('#' + b.getAttribute('data-modal-open'));
      if (m) m.classList.add('open');
    });
  });
  qsa('[data-modal-close], .modal-back').forEach(function (b) {
    b.addEventListener('click', function (e) {
      if (b.classList.contains('modal-back') && e.target !== b) return;
      b.closest('.modal-back').classList.remove('open');
    });
  });

  /* ----- Reel fullscreen player (landing) ----- */
  var reelModal = qs('#reelPlayer');
  if (reelModal) {
    var reelVideo = qs('.ig-modal-video', reelModal);
    var reelTitle = qs('.ig-modal-title', reelModal);
    var reelCreator = qs('.ig-modal-creator', reelModal);

    qsa('[data-video]').forEach(function (tile) {
      tile.addEventListener('click', function () {
        if (!reelVideo) return;
        reelVideo.pause();
        reelVideo.src = tile.getAttribute('data-video') || '';
        if (reelTitle) reelTitle.textContent = tile.getAttribute('data-title') || 'TourChef reel';
        if (reelCreator) reelCreator.textContent = tile.getAttribute('data-creator') || '';
        reelVideo.play().catch(function () { /* autoplay may be blocked */ });
      });
    });

    reelModal.addEventListener('click', function () {
      if (!reelModal.classList.contains('open') && reelVideo) {
        reelVideo.pause();
        reelVideo.removeAttribute('src');
        reelVideo.load();
      }
    });

    var igShare = qs('#igShare');
    if (igShare) {
      igShare.addEventListener('click', function () {
        var url = window.location.href.split('#')[0];
        if (navigator.share) {
          navigator.share({ title: 'TourChef reel', url: url }).catch(function () {});
        } else if (navigator.clipboard) {
          navigator.clipboard.writeText(url).then(function () { toast('Reel link copied.'); });
        } else toast('Copy this page URL to share the reel.');
      });
    }
  }

  document.addEventListener('keydown', function (e) {
    if (e.key !== 'Escape') return;
    qsa('.modal-back.open').forEach(function (m) {
      m.classList.remove('open');
      var v = qs('video', m);
      if (v) { v.pause(); v.removeAttribute('src'); v.load(); }
    });
  });

  /* ----- Demo form submit ----- */
  qsa('form[data-demo]').forEach(function (f) {
    f.addEventListener('submit', function (e) {
      e.preventDefault();
      var btn = qs('button[type="submit"]', f);
      if (btn) {
        btn.classList.add('is-loading');
        setTimeout(function () { btn.classList.remove('is-loading'); toast(f.getAttribute('data-demo')); }, 650);
      } else toast(f.getAttribute('data-demo'));
    });
  });

  /* ----- Search submit: go to catalog with query in URL ----- */
  qsa('form[data-search]').forEach(function (f) {
    f.addEventListener('submit', function (e) {
      e.preventDefault();
      var v = qs('input[data-q]', f);
      location.href = 'masterclasses.html' + (v && v.value ? '?q=' + encodeURIComponent(v.value) : '');
    });
  });

  /* ----- Stepper ----- */
  qsa('[data-stepper]').forEach(function (st) {
    var out = qs('[data-stepper-out]', st);
    var v = parseInt(out && out.textContent, 10) || 1;
    qsa('button', st).forEach(function (b) {
      b.addEventListener('click', function () {
        v = Math.max(1, Math.min(12, v + (b.dataset.step === 'up' ? 1 : -1)));
        if (out) out.textContent = v;
        var priceEl = qs('[data-unit-price]');
        var totalEl = qs('[data-total]');
        if (priceEl && totalEl) {
          var unit = parseInt(priceEl.getAttribute('data-unit-price'), 10);
          totalEl.textContent = new Intl.NumberFormat('uz-UZ').format(unit * v) + " so'm";
        }
      });
    });
  });

  /* ----- Booking widget: session selection ----- */
  qsa('.session-opt').forEach(function (opt) {
    opt.addEventListener('click', function () {
      qsa('.session-opt').forEach(function (o) { o.classList.remove('on'); });
      opt.classList.add('on');
      toast('Sana tanlandi: ' + (qs('.d', opt) || {}).textContent);
    });
  });

  /* ----- Favorite cards ----- */
  qsa('[data-fav]').forEach(function (b) {
    b.addEventListener('click', function () {
      var on = b.classList.toggle('on');
      toast(on ? 'Sevimlilarga qo\'shildi' : 'Sevimlilardan o\'chirildi');
    });
  });

  /* ----- Reels: IntersectionObserver plays visible video ----- */
  var shell = qs('.reels-shell');
  if (shell) {
    var reels = qsa('.reel', shell);
    var videos = qsa('video.rv', shell);
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        var reel = en.target;
        var video = qs('video.rv', reel);
        if (en.isIntersecting && en.intersectionRatio > .6) {
          reels.forEach(function (r) { r.classList.remove('playing'); });
          reel.classList.add('playing');
          if (video) { var p = video.play(); if (p && p.catch) p.catch(function () {}); }
        } else {
          reel.classList.remove('playing');
          if (video) video.pause();
        }
      });
    }, { root: shell, threshold: [0, .6, 1] });
    reels.forEach(function (r) { io.observe(r); });

    /* click video = pause / play */
    qsa('.reel-frame').forEach(function (f) {
      f.addEventListener('click', function (e) {
        if (e.target.closest('button, a')) return;
        var reel = f.closest('.reel');
        var v = qs('video.rv', f);
        var playing = reel.classList.toggle('playing');
        if (v) { if (playing) { var p = v.play(); if (p && p.catch) p.catch(function () {}); } else v.pause(); }
      });
    });
  }

  /* ----- Admin: select all checkbox ----- */
  qsa('[data-checkall]').forEach(function (c) {
    c.addEventListener('change', function () {
      var scope = c.closest('.panel, .acard, table') || document;
      qsa('input[type="checkbox"][data-row]', scope).forEach(function (x) { x.checked = c.checked; });
    });
  });

  /* ----- Dynamic form modals for every "add / create" action ----- */
  var FORMS = {
    'New branch form opened': {
      title: 'Add new branch', sub: 'Create a new location for your restaurant.',
      fields: [
        { label: 'Branch name', ph: 'Osh Markazi · Andijan' },
        { label: 'Parent restaurant', type: 'select', options: ['Osh Markazi', 'Samarkand Kitchen', 'Bukhara Teahouse'] },
        { label: 'Region', type: 'select', options: ['Tashkent Region', 'Samarkand Region', 'Bukhara Region', 'Fergana Region'] },
        { label: 'City', ph: 'Tashkent' },
        { label: 'Address', ph: '14 Navoiy Street' },
        { label: 'Phone', type: 'tel', ph: '+998 90 123 45 67' },
        { label: 'Opening hours', ph: '09:00 – 22:00 daily' },
        { label: 'Manager', type: 'select', options: ['Rustam Karimov', 'Dilshod Khasanov', 'Aziza Nigmatova'] },
        { label: 'Status', type: 'select', options: ['Active', 'Pending approval'] }
      ], submit: 'Create branch', done: 'Branch created (demo)'
    },
    'New masterclass form opened': {
      title: 'New masterclass', sub: 'Build a reusable culinary experience template.',
      fields: [
        { label: 'Title', ph: 'e.g. The Art of Uzbek Plov' },
        { label: 'Description', type: 'textarea', rows: 3, ph: 'What will guests learn?' },
        { label: 'Dish', type: 'select', options: ['Plov', 'Manti', 'Somsa', 'Lagman', 'Shashlik', 'Shurpa', 'Chuchvara'] },
        { label: 'Category', type: 'select', options: ['Main course', 'Dumplings', 'Pastry', 'Grill', 'Soup'] },
        { label: 'Branch', type: 'select', options: ['Osh Markazi · Tashkent', 'Samarkand Kitchen', 'Bukhara Teahouse'] },
        { label: 'Chef / instructor', type: 'select', options: ['Rustam Karimov', 'Sherzod O.', 'Aziza N.'] },
        { label: 'Duration (minutes)', type: 'number', ph: '180' },
        { label: 'Difficulty', type: 'select', options: ['Beginner', 'Intermediate', 'Advanced'] },
        { label: 'Language', type: 'select', options: ['Uzbek', 'Russian', 'English'] },
        { label: 'Minimum participants', type: 'number', ph: '2' },
        { label: 'Maximum participants', type: 'number', ph: '8' },
        { label: 'Price per person (so\'m)', type: 'number', ph: '180000' },
        { label: 'Included services', type: 'textarea', rows: 2, ph: 'Ingredients, apron, tea...' },
        { label: 'Cancellation policy', type: 'textarea', rows: 2, ph: 'Free cancellation up to 24 hours before' }
      ], submit: 'Save as draft', done: 'Masterclass saved as draft (demo)'
    },
    'New session form opened': {
      title: 'New session', sub: 'Schedule a specific event for a masterclass.',
      fields: [
        { label: 'Masterclass', type: 'select', options: ['The Art of Uzbek Plov', 'Hand-folded Manti', 'Somsa from the Tandoor', 'Charcoal Shashlik'] },
        { label: 'Branch', type: 'select', options: ['Osh Markazi · Tashkent', 'Samarkand Kitchen', 'Bukhara Teahouse'] },
        { label: 'Date', type: 'date' },
        { label: 'Start time', type: 'time' },
        { label: 'Duration (minutes)', type: 'number', ph: '180' },
        { label: 'Capacity', type: 'number', ph: '8' },
        { label: 'Minimum participants', type: 'number', ph: '2' },
        { label: 'Price (so\'m)', type: 'number', ph: '180000' },
        { label: 'Booking deadline', type: 'date' },
        { label: 'Cancellation deadline', type: 'date' },
        { label: 'Status', type: 'select', options: ['Scheduled', 'Draft'] }
      ], submit: 'Create session', done: 'Session created (demo)'
    },
    'Invite form opened': {
      title: 'Invite team member', sub: 'Send an invitation to join your restaurant team.',
      fields: [
        { label: 'Full name', ph: 'Aziza Karimova' },
        { label: 'Email', type: 'email', ph: 'teammate@example.com' },
        { label: 'Role', type: 'select', options: ['Branch Manager', 'Chef / Instructor', 'Owner'] },
        { label: 'Branches', type: 'select', options: ['All branches', 'Tashkent', 'Samarkand', 'Bukhara'] },
        { label: 'Personal message', type: 'textarea', rows: 2, ph: 'Welcome to our kitchen!' }
      ], submit: 'Send invitation', done: 'Invitation sent (demo)'
    },
    'New dish form opened': {
      title: 'Add dish', sub: 'Add a dish to the central catalog.',
      fields: [
        { label: 'Dish name', ph: 'e.g. Chuchvara' },
        { label: 'Category', type: 'select', options: ['Main course', 'Dumplings', 'Pastry', 'Soup', 'Grill'] },
        { label: 'Origin region', type: 'select', options: ['Tashkent', 'Samarkand', 'Bukhara', 'Fergana', 'Khiva'] },
        { label: 'Description', type: 'textarea', rows: 3, ph: 'Short cultural description' },
        { label: 'Ingredients', type: 'textarea', rows: 2, ph: 'Flour, meat, broth...' },
        { label: 'Image', type: 'file', accept: 'image/*' },
        { label: 'SEO title', ph: 'Chuchvara — traditional Uzbek dumplings' },
        { label: 'Status', type: 'select', options: ['Active', 'Draft'] }
      ], submit: 'Add dish', done: 'Dish added to catalog (demo)'
    },
    'New location form opened': {
      title: 'Add location', sub: 'Country → Region → City → District hierarchy.',
      fields: [
        { label: 'Country', type: 'select', options: ['Uzbekistan'] },
        { label: 'Region', ph: 'e.g. Tashkent Region' },
        { label: 'City', ph: 'e.g. Chirchiq' },
        { label: 'District', ph: 'e.g. Chirchiq District' },
        { label: 'Coordinates', ph: '41.4200, 69.5600' }
      ], submit: 'Add location', done: 'Location added (demo)'
    },
    'Write a review': {
      title: 'Write a review', sub: 'Only completed and attended experiences can be reviewed.',
      fields: [
        { label: 'Masterclass', type: 'select', options: ['Hand-folded Manti', 'Lagman like Grandma'] },
        { label: 'Rating', type: 'select', options: ['5 — Excellent', '4 — Good', '3 — Average', '2 — Poor', '1 — Terrible'] },
        { label: 'Your review', type: 'textarea', rows: 4, ph: 'Share your experience...' },
        { label: 'Add photos / video', type: 'file', accept: 'image/*,video/*' }
      ], submit: 'Publish review', done: 'Review submitted for moderation (demo)'
    },
    'Upload dialog opened': {
      title: 'Change cover image', sub: 'JPG or PNG · max 10 MB · recommended 1600×600.',
      fields: [{ label: 'Cover image', type: 'file', accept: 'image/*' }, { label: 'Alt text', ph: 'Restaurant interior' }],
      submit: 'Upload cover', done: 'Cover uploaded (demo)'
    },
    'Gallery upload opened': {
      title: 'Add gallery image', sub: 'JPG or PNG · max 10 MB.',
      fields: [{ label: 'Image', type: 'file', accept: 'image/*' }, { label: 'Caption', ph: 'Our tandoor' }],
      submit: 'Add to gallery', done: 'Image added to gallery (demo)'
    },
    'Cancellation request sent': {
      title: 'Cancel booking', sub: 'Free cancellation is available up to 24 hours before the session.',
      fields: [{ label: 'Reason', type: 'select', options: ['Change of plans', 'Illness', 'Found another experience', 'Other'] }, { label: 'Comment', type: 'textarea', rows: 2, ph: 'Optional' }],
      submit: 'Confirm cancellation', done: 'Cancellation request sent (demo)'
    },
    'Report form opened': {
      title: 'Report content', sub: 'Reports are reviewed by the platform team.',
      fields: [{ label: 'Reason', type: 'select', options: ['Inappropriate language', 'Spam or advertisement', 'Incorrect information', 'Other'] }, { label: 'Details', type: 'textarea', rows: 3, ph: 'Tell us what happened' }],
      submit: 'Send report', done: 'Report submitted (demo)'
    },
    'Rejection dialog opened': {
      title: 'Reject submission', sub: 'The owner will see the reason you provide.',
      fields: [{ label: 'Reason', type: 'select', options: ['Incomplete documents', 'Content violates policy', 'Incorrect information', 'Other'] }, { label: 'Explanation', type: 'textarea', rows: 3, ph: 'Explain what needs to be fixed' }],
      submit: 'Reject', done: 'Submission rejected with reason (demo)'
    },
    'Refund dialog opened': {
      title: 'Issue refund', sub: 'Refunds must follow the platform refund policy.',
      fields: [{ label: 'Booking ID', ph: 'TC-0000' }, { label: 'Amount (so\'m)', type: 'number', ph: '360000' }, { label: 'Reason', type: 'select', options: ['Guest request', 'Session cancelled', 'Payment error', 'Dispute resolution'] }, { label: 'Note', type: 'textarea', rows: 2, ph: 'Internal note' }],
      submit: 'Issue refund', done: 'Refund initiated (demo)'
    },
    'Response editor opened': {
      title: 'Respond to review', sub: 'Your response will be visible to the guest and other travelers.',
      fields: [{ label: 'Your response', type: 'textarea', rows: 4, ph: 'Thank you for your feedback...' }],
      submit: 'Publish response', done: 'Response published (demo)'
    },
    'Editor opened': {
      title: 'Edit item', sub: 'Update the details and save your changes.',
      fields: [{ label: 'Title', ph: 'Current title' }, { label: 'Description', type: 'textarea', rows: 3, ph: 'Current description' }, { label: 'Status', type: 'select', options: ['Active', 'Draft', 'Hidden'] }],
      submit: 'Save changes', done: 'Changes saved (demo)'
    },
    'Hide dialog opened': {
      title: 'Hide item', sub: 'Hidden items are no longer visible to the public.',
      fields: [{ label: 'Reason', type: 'select', options: ['Low quality content', 'Outdated information', 'Policy violation', 'Other'] }, { label: 'Note', type: 'textarea', rows: 2, ph: 'Internal note' }],
      submit: 'Hide item', done: 'Item hidden (demo)'
    }
  };
  /* aliases used by other pages */
  FORMS['New review form opened'] = FORMS['Write a review'];
  FORMS['Media upload opened'] = FORMS['Gallery upload opened'];
  FORMS['Image picker opened'] = FORMS['Upload dialog opened'];

  function fieldHTML(f) {
    var input;
    if (f.type === 'select') {
      input = '<select class="select">' + (f.options || []).map(function (o) { return '<option>' + o + '</option>'; }).join('') + '</select>';
    } else if (f.type === 'textarea') {
      input = '<textarea class="textarea" rows="' + (f.rows || 3) + '" placeholder="' + (f.ph || '') + '"></textarea>';
    } else if (f.type === 'file') {
      input = '<input type="file" class="inp" accept="' + (f.accept || '') + '">';
    } else {
      input = '<input class="inp" type="' + (f.type || 'text') + '" placeholder="' + (f.ph || '') + '">';
    }
    return '<div class="field"><label>' + f.label + '</label>' + input + (f.hint ? '<span class="hint">' + f.hint + '</span>' : '') + '</div>';
  }

  function openForm(action) {
    var cfg = FORMS[action];
    if (!cfg) return false;
    if (qs('#formModal')) qs('#formModal').remove();
    var back = document.createElement('div');
    back.className = 'modal-back open';
    back.id = 'formModal';
    back.innerHTML = '<div class="modal"><h3>' + cfg.title + '</h3>' +
      '<p class="muted" style="font-size:13px">' + (cfg.sub || '') + '</p>' +
      '<form style="display:grid;gap:14px;margin-top:18px">' +
      cfg.fields.map(fieldHTML).join('') +
      '<div class="modal-foot"><button type="button" class="btn btn-outline" data-close>Cancel</button>' +
      '<button type="submit" class="btn btn-primary">' + cfg.submit + '</button></div></form></div>';
    document.body.appendChild(back);
    back.addEventListener('click', function (e) {
      if (e.target === back || e.target.closest('[data-close]')) back.remove();
    });
    qs('form', back).addEventListener('submit', function (e) {
      e.preventDefault();
      var btn = qs('[type="submit"]', this);
      btn.classList.add('is-loading');
      setTimeout(function () { back.remove(); toast(cfg.done); }, 550);
    });
    var first = qs('input, select, textarea', back);
    if (first) first.focus();
    return true;
  }

  /* ----- Buttons: open modal form or show demo toast ----- */
  qsa('[data-action]').forEach(function (b) {
    b.addEventListener('click', function (e) {
      e.preventDefault();
      var action = b.getAttribute('data-action');
      if (openForm(action)) return;
      toast(action);
    });
  });

  /* ----- Calendar day select ----- */
  qsa('.cal-day').forEach(function (d) {
    d.addEventListener('click', function () {
      qsa('.cal-day').forEach(function (x) { x.style.outline = ''; });
      d.style.outline = '2px solid var(--accent)';
      toast('Tanlangan sana: ' + d.textContent.trim().split('\n')[0]);
    });
  });

  /* ----- Active nav based on file name ----- */
  var here = location.pathname.split('/').pop();
  qsa('.main-nav a, .cab-nav a, .admin-nav a, .reels-nav a').forEach(function (a) {
    if (a.getAttribute('href') === here) a.classList.add('on', 'active');
  });
})();
