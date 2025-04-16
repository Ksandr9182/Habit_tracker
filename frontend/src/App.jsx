import React, { useState, useEffect } from "react";
import axios from "axios";
import "./App.css";

const buttonGroups = {
    odno_seroe: ["url('/static/image_bg_01.jpg')"],
    odno_krasnoe: ["url('/static/image_bg_03.jpg')"],
    odno_fiolet: ["url('/static/image_bg_04.jpg')"],
    dva_fiolet: ["url('/static/image_bg_04.jpg')", "url('/static/image_bg_06.jpg')"],
    tree: ["url('/static/image_bg_01.jpg')", "url('/static/image_bg_02.jpg')", "url('/static/image_bg_03.jpg')"],
    dva: ["url('/static/image_bg_01.jpg')", "url('/static/image_bg_03.jpg')"],
    dva_zel: ["url('/static/image_bg_01.jpg')", "url('/static/image_bg_02.jpg')"],
    lkar: ["url('/static/lk.jpg')"],
    nastroenie: ["url('/static/nastroenie_01.jpg')", "url('/static/nastroenie_02.jpg')", "url('/static/nastroenie_03.jpg')", "url('/static/nastroenie_04.jpg')"],
    kofe: ["url('/static/kofe.jpg')"],
    vitamin: ["url('/static/vit.jpg')"],
    golov_bol: ["url('/static/golov_bol.jpg')", "url('/static/parik.jpg')"],
    zub: ["url('/static/zub.jpg')", "url('/static/image_bg_03.jpg')"],
    alko: ["url('/static/alko.jpg')"],
    dush: ["url('/static/dush.jpg')"],
    tualet: ["url('/static/tualet.jpg')"],
    fastfood: ["url('/static/fastfud.jpg')", "url('/static/chikmil.jpg')"],
    e1800: ["url('/static/18.jpg')"],
    hod: ["url('/static/hod.jpg')"],
    muchn: ["url('/static/muchn.jpg')"],
    maz: ["url('/static/maz.jpg')", "url('/static/image_bg_03.jpg')"],
    klava: ["url('/static/klava.jpg')", "url('/static/dacha.jpg')", "url('/static/ortodont.jpg')"],
    rabotastatus: ["url('/static/image_bg_01.jpg')", "url('/static/image_bg_02.jpg')", "url('/static/image_bg_04.jpg')", "url('/static/image_bg_05.jpg')", "url('/static/image_bg_06.jpg')"],
};

const idToGroup = {

    kofe01: "kofe",
    kofe02: "kofe",
    kofe03: "odno_krasnoe",
    xxxm: "odno_fiolet",
    xxxs: "dva_fiolet",
    hotba: "hod",
    lkarnitin: "lkar",
    dush: "dush",
    zub01: "zub",
    zub02: "zub",
    rabotastatus: "rabotastatus",
    rabota: "dva_zel",
    rabotautro: "odno_seroe",
    rabotaden: "odno_seroe",
    rabotavecher: "odno_seroe",
    rabotasum: "tree",
    dlinasna: "tree",
    nastroenieutro: "nastroenie",
    nastroenieden: "nastroenie",
    krasotasna: "nastroenie",
    nachalosna: "dva",
    vitamin: "vitamin",
    golova: "golov_bol",
    alkogol: "alko",
    fastfood: "fastfood",
    e1800: "e1800",
    muchnoe: "muchn",
    klava: "klava",
    maz: "maz",
    tualet: "tualet",
};

const textEnabledIds = {
    ves: true,
    rabota: true,
    rabotautro: true,
    rabotaden: true,
    rabotavecher: true,
    rabotasum: true,
    nachalosna: true,
    dlinasna: true,
};

function TrackerButton({ day, index, id, state, text, onChange, activeMonth }) {
    const [inputText, setInputText] = useState(textEnabledIds[id] ? (text || "") : "");
    const [isEditing, setIsEditing] = useState(false);

    useEffect(() => {
        if (textEnabledIds[id]) {
            setInputText(text || "");
        }
    }, [text, id]);

    const group = idToGroup[id] || "odno_seroe";
    const images = buttonGroups[group];
    const maxStates = images.length + 1;

    const handleClick = () => {
        if (!isEditing) {
            const newState = (state + 1) % maxStates;
            onChange(day, index, id, newState, textEnabledIds[id] ? inputText : "", activeMonth);
        }
    };

    const handleDoubleClick = () => {
        if (textEnabledIds[id]) setIsEditing(true);
    };

    const handleTextChange = (e) => {
        setInputText(e.target.value);
        onChange(day, index, id, state, e.target.value, activeMonth);
    };

    const style = {
        backgroundImage: state > 0 ? images[state - 1] : "",
        backgroundColor: state === 0 ? "#999999" : undefined,
        backgroundSize: "auto",
        backgroundRepeat: "repeat",
        backgroundPosition: "top left",
    };

    return (
        <button
            id={id}
            className="knopka"
            style={style}
            onClick={handleClick}
            onDoubleClick={handleDoubleClick}
        >
            {textEnabledIds[id] ? (
                isEditing ? (
                    <input
                        type="text"
                        value={inputText}
                        onChange={handleTextChange}
                        onBlur={() => setIsEditing(false)}
                        maxLength={50}
                        className="knopka-input"
                    />
                ) : (
                    inputText || ""
                )
            ) : (
                ""
            )}
        </button>
    );
}

function DayRow({ day, buttons, onButtonChange, activeMonth }) {
    const buttonIds = [
        "ves", "hotba", "lkarnitin", "kofe01", "kofe02", "kofe03", "zub01", "zub02", "rabotastatus", "rabota",
        "rabotautro", "rabotaden", "rabotavecher", "rabotasum", "xxxs", "nachalosna",
        "dlinasna", "krasotasna", "xxxm", "nastroenieutro", "nastroenieden", "alkogol",
        "dush", "golova", "vitamin", "fastfood", "e1800", "muchnoe", "klava", "maz", "tualet",
    ];

    return (
        <div className="stroka" data-title="Март">
            {buttons.map((btn, index) => (
                <TrackerButton
                    key={buttonIds[index]}
                    day={day}
                    index={index}
                    id={buttonIds[index]}
                    state={btn.state}
                    text={btn.text}
                    onChange={onButtonChange}
                    activeMonth={activeMonth}
                />
            ))}
        </div>
    );
}

function App() {
    const [trackerData, setTrackerData] = useState({});
    const [activeMonth, setActiveMonth] = useState(new Date().getMonth());
    const today = new Date().getDate();

    useEffect(() => {
        axios.get("http://localhost:8080/trackers")
            .then((response) => {
                console.log("Данные с бэкенда:", response.data);
                setTrackerData(response.data);
            })
            .catch((error) => {
                console.error("Ошибка при загрузке данных:", error);
            });
    }, []);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080/ws');
        ws.onopen = () => console.log('WebSocket подключен');
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            console.log('Получены данные через WebSocket:', data);
            setTrackerData(data);
        };
        ws.onerror = (error) => console.error('Ошибка WebSocket:', error);
        ws.onclose = () => console.log('WebSocket закрыт');
        return () => ws.close();
    }, []);

    const handleButtonChange = (day, index, id, state, text, month) => {
        console.log("Отправляем данные:", { day, index, id, state, text, month });
        const dayAsNumber = parseInt(day, 10);
        axios
            .post("http://localhost:8080/trackers/update", {
                day: dayAsNumber,
                index,
                state,
                text,
                month
            })
            .then((response) => {
                console.log("Ответ от сервера:", response.data);
                // Удаляем локальное обновление состояния, так как данные придут через WebSocket
            })
            .catch((error) => {
                console.error("Ошибка при обновлении:", error);
                if (error.response) {
                    console.log("Ответ сервера:", error.response.data);
                }
            });
    };

    const handleMonthClick = (monthIndex) => {
        setActiveMonth(monthIndex);
        // Отправляем запрос на бэкенд для получения данных за выбранный месяц
        axios
            .post("http://localhost:8080/trackers/month", { month: monthIndex })
            .then((response) => {
                console.log("Данные за месяц с бэкенда:", response.data);
                setTrackerData(response.data); // Обновляем данные на фронтенде
            })
            .catch((error) => {
                console.error("Ошибка при загрузке данных за месяц:", error);
                if (error.response) {
                    console.log("Ответ сервера:", error.response.data);
                }
            });
    };

    const months = [
        'Январь', 'Февраль', 'Март', 'Апрель',
        'Май', 'Июнь', 'Июль', 'Август',
        'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
    ];

    const daysOfWeek = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт','Сб', 'Вс'];

    // Функция для получения дней недели для каждого дня месяца
    const getDaysOfWeekForMonth = (month) => {
        const year = 2025; // Фиксируем год
        const daysInMonth = new Date(year, month + 1, 0).getDate(); // Количество дней в месяце
        const days = [];

        for (let day = 1; day <= 31; day++) {
            if (day <= daysInMonth) {
                const date = new Date(year, month, day);
                const dayOfWeekIndex = date.getDay(); // 0 (Вс), 1 (Пн), ..., 6 (Сб)
                // Корректируем индекс, так как daysOfWeek начинается с "Сб"
                const adjustedIndex = (dayOfWeekIndex + 6) % 7; // Сдвигаем: 0 (Вс) → 1, 6 (Сб) → 0
                days.push(daysOfWeek[adjustedIndex]);
            } else {
                days.push(""); // Пустая строка для дней, которых нет в месяце
            }
        }

        return days;
    };

    // Получаем дни недели для текущего месяца
    const daysOfWeekForMonth = getDaysOfWeekForMonth(activeMonth);

    return (
        <div className="container">
            <div className="container-menu">
                {months.map((month, i) => (
                    <div
                        key={i}
                        className={`mesyac ${activeMonth === i ? 'active' : ''}`}
                        onClick={() => handleMonthClick(i)}
                    >
                        {month}
                    </div>
                ))}
            </div>
            <div className="container-mesyac01">
                {Array.from({ length: 31 }, (_, i) => {
                    const dayLabel = daysOfWeekForMonth[i];
                    const additionalClass = (dayLabel === 'Сб' || dayLabel === 'Вс') ? 'vihodnoiy' : '';
                    return (
                        <div key={i + 1} className={`nedelya ${additionalClass}`}>
                            {dayLabel}
                        </div>
                    );
                })}
            </div>
            <div className="container-mesyac02">
                {Array.from({ length: 31 }, (_, i) => {
                    const dayNumber = String(i + 1).padStart(2, "0");
                    const isToday = dayNumber === String(today).padStart(2, "0") && activeMonth === new Date().getMonth();
                    return (
                        <div
                            key={i + 1}
                            className="data"
                            id={isToday ? "segodnya" : undefined}
                        >
                            {dayNumber}
                        </div>
                    );
                })}
            </div>
            <div className="container-stroki">
                {Object.entries(trackerData).map(([day, buttons]) => (
                    <DayRow
                        key={day}
                        day={day}
                        buttons={buttons}
                        onButtonChange={handleButtonChange}
                        activeMonth={activeMonth}
                    />
                ))}
            </div>
        </div>
    );
}

export default App;