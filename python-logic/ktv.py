from datetime import datetime, timezone, timedelta


class KTV_TIME:
    ktv_int_to_char = {
        0: "𝋀", 1: "𝋁", 2: "𝋂", 3: "𝋃", 4: "𝋄",
        5: "𝋅", 6: "𝋆", 7: "𝋇", 8: "𝋈", 9: "𝋉",
        10: "𝋊", 11: "𝋋", 12: "𝋌", 13: "𝋍", 14: "𝋎",
        15: "𝋏", 16: "𝋐", 17: "𝋑", 18: "𝋒", 19: "𝋓"
    }

    ktv_char_to_int = {v: k for k, v in ktv_int_to_char.items()}

    def __init__(self, tz=int):
        if tz >= 24 or tz <= -24:
            raise Exception("bad time zone, straight to jail")
        ikarraq, mein, tick, kick = self.get_ktv_characters(timezone(timedelta(hours=tz)))
        self.ikarraq = ikarraq
        self.mein = mein
        self.tick = tick
        self.kick = kick

    def get_ktv_integer(self, tz) -> tuple[int, int, int, int]:
        now = datetime.now(tz)

        seconds = now.hour * 3600 + now.minute * 60 + now.second + now.microsecond / 1_000_000
        fraction_of_the_day = seconds / 86400

        ikarraq = int(fraction_of_the_day * 20)
        remainder = (fraction_of_the_day * 20) % 1

        mein = int(remainder * 20)
        remainder = (remainder * 20) % 1

        tick = int(remainder * 20)
        remainder = (remainder * 20) % 1

        kick = int(remainder * 20)

        return (ikarraq, mein, tick, kick)

    def get_ktv_characters(self, tz=timezone.utc) -> tuple[str, str, str, str]:
        i, m, c, k = self.get_ktv_integer(tz)
        return (
            KTV_TIME.ktv_int_to_char[i],
            KTV_TIME.ktv_int_to_char[m],
            KTV_TIME.ktv_int_to_char[c],
            KTV_TIME.ktv_int_to_char[k]
        )

    @staticmethod
    def imperial_to_ktv(hh: int, mm: int = 0, ss: int = 0, ms: int = 0, ampm: str = None) -> tuple[str, str, str, str]:
        """
        Converts an imperial time (12-hour or 24-hour) into Kaktovik time characters.
        """
        # Handle AM/PM conversion only if hh <= 12
        if hh <= 12 and ampm:
            ampm = ampm.lower()
            if ampm == "pm" and hh != 12:
                hh += 12
            elif ampm == "am" and hh == 12:
                hh = 0
        elif hh > 23:
            raise Exception("Hour must be between 0-23 for 24-hour format or 1-12 with AM/PM")

        # Total seconds in the day
        total_seconds = hh * 3600 + mm * 60 + ss + (ms / 1000.0)
        fraction_of_the_day = total_seconds / 86400

        # Break into base-20
        ikarraq = int(fraction_of_the_day * 20)
        remainder = (fraction_of_the_day * 20) % 1

        mein = int(remainder * 20)
        remainder = (remainder * 20) % 1

        tick = int(remainder * 20)
        remainder = (remainder * 20) % 1

        kick = int(remainder * 20)

        return (
            ikarraq,
            mein,
            tick,
            kick
        )

    @staticmethod
    def ktv_to_imperial(ikarraq, mein, tick, kick) -> tuple[int, int, int, int]:
        """
        Converts a Kaktovik time (characters or integers) into exact imperial time.
        Returns (hour, minute, second, millisecond)
        """
        # Convert characters to integers if necessary
        def to_int(v: str | int) -> int:
            if v.isnumeric():
                return int(v)
            if isinstance(v, str):
                return KTV_TIME.ktv_char_to_int[v.strip()]
            return int(v)

        i = to_int(ikarraq)
        m = to_int(mein)
        t = to_int(tick)
        k = to_int(kick)

        # Total fraction of the day
        fraction_of_the_day = (i / 20) + (m / (20 ** 2)) + (t / (20 ** 3)) + (k / (20 ** 4))
        total_seconds = fraction_of_the_day * 86400

        hh = int(total_seconds // 3600)
        total_seconds %= 3600

        mm = int(total_seconds // 60)
        total_seconds %= 60

        ss = int(total_seconds)
        ms = int(round((total_seconds - ss) * 1000))

        return (hh, mm, ss, ms)
