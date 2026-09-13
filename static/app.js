const HOLD_SECONDS = 120;

const MOVIES = [
    { id: "cidade-de-deus", title: "Cidade de Deus", rows: 5, seatsPerRow: 8 },
    { id: "bacurau", title: "Bacurau", rows: 4, seatsPerRow: 6 },
];

const userId = crypto.randomUUID().slice(0, 12);
document.getElementById("userIdLabel").textContent = userId;

const baseUrlEl = document.getElementById("baseUrl");
const moviesEl = document.getElementById("movies");
const seatGridEl = document.getElementById("seatGrid");
const checkoutEl = document.getElementById("checkout");
const statusMsgEl = document.getElementById("statusMsg");

let selectedMovie = MOVIES[1];
let confirmedSeats = new Set();
let hold = null; // { seatId, expiresAt }
let timerInterval = null;

function rowLabel(index) {
    return String.fromCharCode(65 + index);
}

function renderMovies() {
    moviesEl.innerHTML = "";
    MOVIES.forEach((movie) => {
        const card = document.createElement("div");
        card.className = `movie-card${movie.id === selectedMovie.id ? " selected" : ""}`;
        card.innerHTML = `<h3>${movie.title}</h3><p>${movie.rows} fileiras × ${movie.seatsPerRow} assentos</p>`;
        card.addEventListener("click", () => selectMovie(movie));
        moviesEl.appendChild(card);
    });
}

function selectMovie(movie) {
    if (movie.id === selectedMovie.id) return;
    if (hold) releaseHold(false);
    selectedMovie = movie;
    confirmedSeats = new Set();
    renderMovies();
    renderGrid();
    fetchSeats();
}

function renderGrid() {
    seatGridEl.innerHTML = "";
    for (let r = 0; r < selectedMovie.rows; r++) {
        const rowEl = document.createElement("div");
        rowEl.className = "seat-row";

        const leftLabel = document.createElement("div");
        leftLabel.className = "row-label";
        leftLabel.textContent = rowLabel(r);
        rowEl.appendChild(leftLabel);

        for (let s = 1; s <= selectedMovie.seatsPerRow; s++) {
            const seatId = `${rowLabel(r)}${s}`;
            const seatEl = document.createElement("div");
            seatEl.className = "seat";
            seatEl.textContent = s;
            seatEl.dataset.seatId = seatId;

            if (confirmedSeats.has(seatId)) {
                seatEl.classList.add("seat--confirmed");
            } else if (hold && hold.seatId === seatId) {
                seatEl.classList.add("seat--hold-mine");
            } else {
                seatEl.addEventListener("click", () => startHold(seatId));
            }

            rowEl.appendChild(seatEl);
        }

        const rightLabel = document.createElement("div");
        rightLabel.className = "row-label";
        rightLabel.textContent = rowLabel(r);
        rowEl.appendChild(rightLabel);

        seatGridEl.appendChild(rowEl);
    }
}

function startHold(seatId) {
    if (hold) return;
    hold = { seatId, expiresAt: Date.now() + HOLD_SECONDS * 1000 };
    renderGrid();
    renderCheckout();
    startTimer();
}

function releaseHold(rerender = true) {
    hold = null;
    clearTimerInterval();
    checkoutEl.innerHTML = "";
    if (rerender) renderGrid();
}

async function confirmHold() {
    if (!hold) return;
    const seatId = hold.seatId;
    setStatus("Confirmando...", "");
    try {
        const resp = await fetch(`${baseUrlEl.value}/reservas`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                filme_id: selectedMovie.id,
                assento_id: seatId,
                usuario_id: userId,
            }),
        });
        const data = await resp.json();
        if (resp.ok) {
            setStatus(`Assento ${seatId} confirmado!`, "success");
            releaseHold(false);
            await fetchSeats();
        } else {
            setStatus(
                `Erro ao confirmar ${seatId}: ${data.erro || resp.status}`,
                "error",
            );
            releaseHold(false);
            await fetchSeats();
        }
    } catch (e) {
        setStatus(`Falha na requisição: ${e.message}`, "error");
    }
}

function setStatus(msg, kind) {
    statusMsgEl.textContent = msg;
    statusMsgEl.className = `status-msg ${kind || ""}`;
}

function renderCheckout() {
    if (!hold) {
        checkoutEl.innerHTML = "";
        return;
    }
    checkoutEl.innerHTML = `
        <div class="checkout">
          <h3>Checkout</h3>
          <div class="checkout-info">
            <span>Filme: ${selectedMovie.title}</span>
            <span>Assento: ${hold.seatId}</span>
          </div>
          <div class="timer" id="timerLabel"></div>
          <div class="checkout-buttons">
            <button class="btn btn--confirm" id="btnConfirm">Confirmar</button>
            <button class="btn btn--release" id="btnRelease">Cancelar</button>
          </div>
        </div>
      `;
    document.getElementById("btnConfirm").addEventListener("click", confirmHold);
    document.getElementById("btnRelease").addEventListener("click", () => {
        setStatus("Hold cancelado.", "");
        releaseHold();
    });
    updateTimer();
}

function startTimer() {
    clearTimerInterval();
    renderCheckout();
    timerInterval = setInterval(updateTimer, 1000);
}

function updateTimer() {
    if (!hold) return;
    const remainingMs = hold.expiresAt - Date.now();
    if (remainingMs <= 0) {
        setStatus("Tempo esgotado, assento liberado.", "error");
        releaseHold();
        return;
    }
    const totalSeconds = Math.ceil(remainingMs / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    const label = document.getElementById("timerLabel");
    if (label) {
        label.textContent = `${minutes}:${String(seconds).padStart(2, "0")}`;
        label.className = `timer${totalSeconds <= 20 ? " urgent" : ""}`;
    }
}

function clearTimerInterval() {
    if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
    }
}

async function fetchSeats() {
    try {
        const resp = await fetch(
            `${baseUrlEl.value}/filmes/${encodeURIComponent(selectedMovie.id)}/reservas`,
        );
        const lista = resp.ok ? await resp.json() : [];
        confirmedSeats = new Set((lista || []).map((r) => r.AssentoID));
        renderGrid();
    } catch (e) {
        setStatus(`Falha ao carregar assentos: ${e.message}`, "error");
    }
}

baseUrlEl.addEventListener("change", fetchSeats);

renderMovies();
renderGrid();
fetchSeats();
setInterval(fetchSeats, 2000);
