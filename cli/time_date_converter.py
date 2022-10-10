def function(seconds_total):

    def text(time_value):
        if time_value == "seconds_total":
            if (seconds_total > 1) or (seconds_total == 0):
                return "seconds"
            else:
                return "second"

        if time_value == "seconds_remainder":
            if (seconds_remainder > 1) or (seconds_remainder == 0):
                return "seconds"
            else:
                return "second"

        if time_value == "minutes_total":
            if (minutes_total > 1) or (minutes_total == 0):
                return "minutes"
            else:
                return "minute"

        if time_value == "minutes_remainder":
            if (minutes_remainder > 1) or (minutes_remainder == 0):
                return "minutes"
            else:
                return "minute"

        if time_value == "hours_total":
            if (hours_total > 1) or (hours_total == 0):
                return "hours"
            else:
                return "hour"

        if time_value == "hours_remainder":
            if (hours_remainder > 1) or (hours_remainder == 0):
                return "hours"
            else:
                return "hour"

        if time_value == "days_total":
            if (days_total > 1) or (days_total == 0):
                return "days"
            else:
                return "day"

        if time_value == "days_remainder":
            if (days_remainder > 1) or (days_remainder == 0):
                return "days"
            else:
                return "day"

    seconds_total = round(seconds_total)
    weeks_total = int(seconds_total / 604800)
    days_total = int(seconds_total / 86400)
    hours_total = int(seconds_total / 3600)
    minutes_total = round(seconds_total / 60)

    days_remainder = days_total % 7
    hours_remainder = hours_total % 24
    minutes_remainder = minutes_total % 60
    seconds_remainder = seconds_total % 60

    if (seconds_total >= 60) and (minutes_total < 60):
        return f'{minutes_total} {text("minutes_total")} {seconds_remainder} {text("seconds_remainder")}'
    elif (minutes_total >= 60) and (hours_total < 24):
        return f'{hours_total} {text("hours_total")} {minutes_remainder} {text("minutes_remainder")}'
    elif hours_total >= 24:
        return f'{days_total} {text("days_total")} {hours_remainder} {text("hours_remainder")} {minutes_remainder} {text("minutes_remainder")}'
    else:
        return f'{seconds_total} {text("seconds_total")}'